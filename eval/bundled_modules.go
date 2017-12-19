package eval

var bundledModules = map[string]string{
	"epm": `-data-dir = ~/.elvish
-repo = $-data-dir/pkgindex
-installed = $-data-dir/epm-installed
-repo-url = https://github.com/elves/pkgindex

fn -info [text]{
    print (edit:styled '=> ' green)
    echo $text
}

fn -error [text]{
    print (edit:styled '=> ' red)
    echo $text
}

fn update {
    if (-is-dir $-repo) {
        -info 'Updating epm repo...'
        git -C $-repo pull
    } else {
        -info 'Cloning epm repo...'
        git clone $-repo-url $-repo
    }
}

fn add-installed [pkg]{
    echo $pkg >> $-installed
}

fn -install-one [pkg]{
    dest = $-data-dir/lib/$pkg
    if ?(test -e $dest) {
        -error 'Package '$pkg' already exists locally.'
        return
    }
    metafile = $-repo/pkg/$pkg
    if (not ?(test -f $metafile)) {
        -error 'Package '$pkg' not found. Try epm:update?'
        return
    }
    meta = (cat $metafile | from-json)
    url desc = $meta[url description]
    -info 'Installing package '$pkg': '$desc
    git clone $url $dest
    add-installed $pkg
}

fn install [@pkgs]{
    if (eq $pkgs []) {
        -error 'Must specify at least one package.'
        return
    }
    for pkg $pkgs {
        -install-one $pkg
    }
}

fn installed {
    if ?(test -f $-installed) {
        cat $-installed
    }
}

fn -upgrade-one [pkg]{
    dest = $-data-dir/lib/$pkg
    if (not ?(test -d $dest)) {
        -error 'Package '$pkg' not installed locally.'
        return
    }
    -info 'Upgrading package '$pkg
    git -C $dest pull
}

fn upgrade [@pkgs]{
    if (eq $pkgs []) {
        pkgs = [(installed)]
        -info 'Upgrading all installed packages'
    }
    for pkg $pkgs {
        -upgrade-one $pkg
    }
}

fn -uninstall-one [pkg]{
    installed-pkgs = [(installed)]
    if (not (has-value $installed-pkgs $pkg)) {
        -error 'Package '$pkg' is not registered as installed.'
        return
    }
    dest = $-data-dir/lib/$pkg
    if (not ?(test -d $dest)) {
        -error 'Package '$pkg' does not exist locally.'
        return
    }
    -info 'Removing package '$pkg
    rm -rf $dest
    # issue #486
    {
        for installed $installed-pkgs {
            if (not-eq $installed $pkg) {
                echo $installed
            }
        }
    } > $-installed
}

fn uninstall [@pkgs]{
    if (eq $pkgs []) {
        -error 'Must specify at least one package.'
        return
    }
    for pkg $pkgs {
        -uninstall-one $pkg
    }
}
`,
	"narrow": `# Implementation of location, history and lastcmd mode using the new
# -narrow-read mode. One advantage of this is that it allows the
# execution of arbitrary hooks before or after each mode.
#
# Usage:
#   use narrow
#   narrow:bind-trigger-keys
#
# narrow:bind-trigger-keys binds keys for location, history and lastcmd
# modes. Without options, it uses the default bindings (same as the
# default bindings for edit:location, edit:history and edit:lastcmd),
# but different keys can be specified with the options. To disable a
# binding, specify its key as "".
# Example:
#   narrow:bind-trigger-keys &location=Alt-l &lastcmd="" 

# Hooks
# Each hook variable is an array which must contain lambdas, all of
# which will be executed in sequence before and after the
# corresponding mode.
# Example (list the new directory after switching to it in location mode):
#    narrow:after-location = [ $@narrow:after-location { edit:insert-at-dot "ls"; edit:smart-enter } ]
before-location = []
after-location = []
before-history = []
after-history = []
before-lastcmd = []
after-lastcmd = []

fn location {
  for hook $before-location { $hook }
  candidates = [(dir-history | each [arg]{
	      score = (splits . $arg[score] | take 1)
        put [
          &content=$arg[path]
          &display=$score" "$arg[path]
	      ]
  })]

  edit:-narrow-read {
    put $@candidates
  } [arg]{
    cd $arg[content]
    for hook $after-location { $hook }
  } &modeline="[narrow] Location " &ignore-case=$true
}

fn history {
  for hook $before-history { $hook }
  candidates = [(edit:command-history | each [arg]{
        put [
	        &content=$arg[cmd]
	        &display=$arg[id]" "(replaces "\t" " " (replaces "\n" " " $arg[cmd]))
        ]
  })]

  edit:-narrow-read {
    put $@candidates
  } [arg]{
    edit:replace-input $arg[content]
    for hook $after-history { $hook }
  } &modeline="[narrow] History " &keep-bottom=$true &ignore-case=$true
}

fn lastcmd {
  for hook $before-lastcmd { $hook }
  last = (edit:command-history -1)
  cmd = [
    &content=$last[cmd]
    &display="M-1 "$last[cmd]
	  &filter-text=""
  ]
  index = 0
  candidates = [$cmd ( edit:wordify $last[cmd] | each [arg]{
	      put [
          &content=$arg
          &display=$index" "$arg
          &filter-text=$index
	      ]
	      index = (+ $index 1)
  })]
  edit:-narrow-read {
    put $@candidates
  } [arg]{
    edit:insert-at-dot $arg[content]
    for hook $after-lastcmd { $hook }
  } &modeline="[narrow] Lastcmd " &auto-commit=$true &bindings=[&M-1={ edit:narrow:accept-close }] &ignore-case=$true
}


fn -bind-insert [k f]{
  edit:insert:binding[$k] = $f
}

fn -bind [k f]{
  edit:narrow:binding[$k] = $f
}

# Bind keys for location, history and lastcmd modes. Without
# options, it uses the default bindings, but different keys
# can be specified with the options. To disable a binding,
# specify its key as "".
# Example:
#   narrow:bind-trigger-keys &location=Alt-l &lastcmd=""
fn bind-trigger-keys [&location=C-l &history=C-r &lastcmd=M-1]{
  if (not-eq $location "") { -bind-insert $location narrow:location }
  if (not-eq $history "")  { -bind-insert $history  narrow:history }
  if (not-eq $lastcmd "")  { -bind-insert $lastcmd  narrow:lastcmd }
}

# Set up some default useful bindings for narrow mode
-bind Up        $edit:narrow:up~
-bind PageUp    $edit:narrow:page-up~
-bind Down      $edit:narrow:down~
-bind PageDown  $edit:narrow:page-down~
-bind Tab       $edit:narrow:down-cycle~
-bind S-Tab     $edit:narrow:up-cycle~
-bind Backspace $edit:narrow:backspace~
-bind Enter     $edit:narrow:accept-close~
-bind M-Enter   $edit:narrow:accept~
-bind default   $edit:narrow:default~
-bind "C-["     $edit:insert:start~
-bind C-G       $edit:narrow:toggle-ignore-case~
-bind C-D       $edit:narrow:toggle-ignore-duplication~
`,
	"readline-binding": `b=[k f]{ edit:insert:binding[$k] = $f } {
    $b Ctrl-A $edit:move-dot-sol~
    $b Ctrl-B $edit:move-dot-left~
    $b Ctrl-D {
        if (> (count $edit:current-command) 0) {
            edit:kill-rune-right
        } else {
            edit:return-eof
        }
    }
    $b Ctrl-E $edit:move-dot-eol~
    $b Ctrl-F $edit:move-dot-right~
    $b Ctrl-H $edit:kill-rune-left~
    $b Ctrl-L { clear > /dev/tty }
    $b Ctrl-N $edit:end-of-history~
    # TODO: ^O
    $b Ctrl-P $edit:history:start~
    # TODO: ^S ^T ^X family ^Y ^_
    $b Alt-b  $edit:move-dot-left-word~
    # TODO Alt-c Alt-d
    $b Alt-f  $edit:move-dot-right-word~
    # TODO Alt-l Alt-r Alt-u

    # Ctrl-N and Ctrl-L occupied by readline binding, $b to Alt- instead.
    $b Alt-n $edit:navigation:start~
    $b Alt-l $edit:location:start~
}

b=[k f]{ edit:completion:binding[$k] = $f } {
    $b Ctrl-B $edit:completion:left~
    $b Ctrl-F $edit:completion:right~
    $b Ctrl-N $edit:completion:down~
    $b Ctrl-P $edit:completion:up~
    $b Alt-f  $edit:completion:trigger-filter~
}

b=[k f]{ edit:navigation:binding[$k] = $f } {
    $b Ctrl-B $edit:navigation:left~
    $b Ctrl-F $edit:navigation:right~
    $b Ctrl-N $edit:navigation:down~
    $b Ctrl-P $edit:navigation:up~
    $b Alt-f  $edit:navigation:trigger-filter~
}

b=[k f]{ edit:history:binding[$k] = $f } {
    $b Ctrl-N $edit:history:down-or-quit~
    $b Ctrl-P $edit:history:up~
    $b Ctrl-G $edit:insert:start~
}

b=[k f]{ edit:listing:binding[$k] = $f } {
    $b Ctrl-N $edit:listing:down~
    $b Ctrl-P $edit:listing:up~
    $b Ctrl-V $edit:listing:page-down~
    $b Alt-v  $edit:listing:page-up~
    $b Ctrl-G $edit:insert:start~
}

b=[k f]{ edit:histlist:binding[$k] = $f } {
    $b Alt-g $edit:histlist:toggle-case-sensitivity~
    $b Alt-d $edit:histlist:toggle-dedup~
}
`,
}
