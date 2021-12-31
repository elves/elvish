use flag
use re
use str

var prefix = src.elv.sh/

fn keep-if {|p| each {|x| if ($p $x) { put $x }} }
fn get {|x k def| if (has-key $x $k) { put $x[$k] } else { put $def } }
fn get-cluster {|x| put (re:find '^'(re:quote $prefix)'[^/]+/[^/]+' $x)[text] }
fn node {|x| put '"'(str:trim-prefix $x $prefix)'"' }

fn main {|&merge-clusters=$false|
  var imports-of = [&]
  var q = [$prefix''cmd/elvish]
  var seen = [&q[0]=$true]
  var clusters = [&]

  while (not-eq $q []) {
    var next-q = []
    for pkg $q {
      var c = (get-cluster $pkg)
      set clusters[$c] = [(all (get $clusters $c [])) $pkg]
      var @imports = (
        go list -json $pkg |
          all (get (from-json) Imports []) |
          keep-if {|x| str:has-prefix $x $prefix})
      set imports-of[$pkg] = $imports

      var @new-pkgs = (all $imports | keep-if {|x|
        not (has-key $seen $x)
        set seen[$x] = $true
      })
      set @next-q = (all $next-q) (all $new-pkgs)
    }
    set q = $next-q
  }

  echo 'strict digraph imports {'
  echo '  rankdir = LR;'
  echo '  node [shape = box, width = 1.5];'
  echo '  splines = ortho;'
  echo '  nodesep = 0.1;'
  if $merge-clusters {
    for pkg [(keys $imports-of)] {
      for import $imports-of[$pkg] {
        var src = (get-cluster $pkg)
        var dst = (get-cluster $import)
        if (not-eq $src $dst) {
          echo '  '(node $src)' -> '(node $dst)';'
        }
      }
    }
  } else {
    var cluster-seq = 0
    for c [(keys $clusters)] {
      var pkgs = $clusters[$c]
      if (<= (count $pkgs) 1) { continue }
      echo '  subgraph cluster'$cluster-seq' {'
      echo '    style = filled;'
      echo '    color = lightgrey;'
      for pkg $clusters[$c] {
        echo '    '(node $pkg)';'
      }
      echo '  }'
      set cluster-seq = (+ $cluster-seq 1)
    }
    for pkg [(keys $imports-of)] {
      for import $imports-of[$pkg] {
        echo '  '(node $pkg)' -> '(node $import)';'
      }
    }
  }
  echo '}'
}

flag:call $main~ $args
