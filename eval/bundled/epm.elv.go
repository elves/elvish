package bundled

const epmElv = `# Verbosity configuration
debug-mode = $false

# Configuration for common domains
-default-domain-config = [
  &"github.com"= [
    &method= git
    &protocol= https
    &levels= 2
  ]
  &"bitbucket.org"= [
    &method= git
    &protocol= https
    &levels= 2
  ]
  &"gitlab.com"= [
    &method= git
    &protocol= https
    &levels= 2
  ]
]

# Internal configuration
-data-dir = ~/.elvish
-lib-dir = $-data-dir/lib

# Runtime state - records get copied from -default-domain-config
# or read from epm-domain.cfg files as needed
-domain-config = [&]

# General utility functions

fn -debug [text]{
  if $debug-mode {
    print (edit:styled '=> ' blue)
    echo $text
  }
}

fn -info [text]{
  print (edit:styled '=> ' green)
  echo $text
}

fn -warn [text]{
  print (edit:styled '=> ' yellow)
  echo $text
}

fn -error [text]{
  print (edit:styled '=> ' red)
  echo $text
}

fn dest [pkg]{
  put $-lib-dir/$pkg
}

fn is-installed [pkg]{
  put ?(test -e (dest $pkg))
}

fn -package-domain [pkg]{
  splits / $pkg | take 1
}

fn -package-without-domain [pkg]{
  joins / [(splits / $pkg | drop 1)]
}

# Known method handlers. Each entry is indexed by method name (the
# value of the "method" key in the domain configs), and must contain
# two keys: install and upgrade, each one must be a closure that
# received two arguments: package name and the domain config entry
-method-handler = [
  &git= [
    &src= [pkg dom-cfg]{
      put $dom-cfg[protocol]"://"$pkg
    }

    &install= [pkg dom-cfg]{
      dest = (dest $pkg)
      -info "Installing "$pkg
      mkdir -p $dest
      git clone ($-method-handler[git][src] $pkg $dom-cfg) $dest
    }

    &upgrade= [pkg dom-cfg]{
      dest = (dest $pkg)
      -info "Updating "$pkg
      git -C $dest pull
    }
  ]

  &rsync= [
    &src= [pkg dom-cfg]{
      put $dom-cfg[location]/(-package-without-domain $pkg)/
    }

    &install= [pkg dom-cfg]{
      dest = (dest $pkg)
      pkgd = (-package-without-domain $pkg)
      -info "Installing "$pkg
      rsync -av ($-method-handler[rsync][src] $pkg $dom-cfg) $dest
    }

    &upgrade= [pkg dom-cfg]{
      dest = (dest $pkg)
      pkgd = (-package-without-domain $pkg)
      if (not (is-installed $pkg)) {
        -error "Package "$pkg" is not installed."
        return
      }
      -info "Updating "$pkg
      rsync -av ($-method-handler[rsync][src] $pkg $dom-cfg) $dest
    }
  ]
]

# Return the filename of the domain config file for the given domain
# (regardless of whether it exists)
fn -domain-config-file [dom]{
  put $-lib-dir/$dom/epm-domain.cfg
}

# Return the filename of the metadata file for the given package
# (regardless of whether it exists)
fn -package-metadata-file [pkg]{
  put (dest $pkg)/metadata.json
}

# Read the domain config file for a given domain. If the file does not
# exist but we have a built-in definition, then we create the file
# with the default.
fn -read-domain-config [dom]{
  cfgfile = (-domain-config-file $dom)
  # Only read config if it hasn't been loaded already
  if (not (has-key $-domain-config $dom)) {
    if ?(test -f $cfgfile) {
      # If the config file exists, read it...
      -domain-config[$dom] = (cat $cfgfile | from-json)
      -debug "Read domain config for "$dom": "(to-string $-domain-config[$dom])
    } else {
      # ...otherwise check if we have a default config for the domain, and save it
      if (has-key $-default-domain-config $dom) {
        -domain-config[$dom] = $-default-domain-config[$dom]
        -debug "No existing config for "$dom", using the default: "(to-string $-domain-config[$dom])
        mkdir -p (dirname $cfgfile)
        put $-domain-config[$dom] | to-json > $cfgfile
      } else {
        fail "No existing config for "$dom" and no default available. Please create config file "(tilde-abbr $cfgfile)" by hand"
      }
    }
  }
}

# Return the method by which a package is installed
fn -package-method [pkg]{
  dom = (-package-domain $pkg)
  -read-domain-config $dom
  if (has-key $-domain-config $dom) {
    cfg = $-domain-config[$dom]
    put $cfg[method]
  } else {
    put $false
  }
}

# Invoke package operations defined in $-method-handler above
fn -package-op [pkg what]{
  dom = (-package-domain $pkg)
  -read-domain-config $dom
  if (has-key $-domain-config $dom) {
    cfg = $-domain-config[$dom]
    method = $cfg[method]
    if (has-key $-method-handler $method) {
      if (has-key $-method-handler[$method] $what) {
        $-method-handler[$method][$what] $pkg $cfg
      } else {
        fail "No handler for '"$what"' defined in method '"$method"'"
      }
    } else {
      fail "No handler defined for method '"$method"', specified in in config file "(-domain-config-file $dom)
    }
  }
}

# Read and parse the package metadata, if it exists
fn metadata [pkg]{
  res = [&]
  mdata = (-package-metadata-file $pkg)
  if (and (is-installed $pkg) ?(test -f $mdata)) {
    res = (cat $mdata | from-json)
  }
  put $res
}

# Print out information about a package
fn query [pkg]{
  data = (metadata $pkg)
  echo (edit:styled "Package "$pkg cyan)
  echo (edit:styled "Source:" blue) (-package-method $pkg) (-package-op $pkg src)
  if (is-installed $pkg) {
    echo (edit:styled "Installed at "(dest $pkg) green)
  } else {
    echo (edit:styled "Not installed" red)
  }
  keys $data | each [key]{
    echo (edit:styled $key":" blue) $data[$key]
  }
}

# Uninstall a single package by removing its directory
fn -uninstall-package [pkg]{
  if (not (is-installed $pkg)) {
    -error "Package "$pkg" is not installed."
    return
  }
  dest = (dest $pkg)
  -info "Removing package "$pkg
  rm -rf $dest
}

######################################################################
# Main user-facing functions

# List installed packages
fn installed {
  e:ls $-lib-dir | each [dom]{
    if ?(test -f (-domain-config-file $dom)) {
      -read-domain-config $dom
      lvl = $-domain-config[$dom][levels]
      find $-lib-dir/$dom -type d -depth $lvl | each [pkg]{
        replaces $-lib-dir/ "" $pkg
      }
    }
  }
}

# Install and upgrade are method-specific, so we call the
# corresponding functions using -package-op
fn install [&silent-if-installed=$false @pkgs]{
  for pkg $pkgs {
    if (is-installed $pkg) {
      if (not $silent-if-installed) {
        -info "Package "$pkg" is already installed."
      }
    } else {
      -package-op $pkg install
    }
  }
}

fn upgrade [@pkgs]{
  if (eq $pkgs []) {
    pkgs = [(installed)]
    -info 'Upgrading all installed packages'
  }
  for pkg $pkgs {
    if (not (is-installed $pkg)) {
      -error "Package "$pkg" is not installed."
    } else {
      -package-op $pkg upgrade
    }
  }
}

# Uninstall is the same for everyone, just remove the directory
fn uninstall [@pkgs]{
  if (eq $pkgs []) {
    fail 'Must specify at least one package.'
    return
  }
  for pkg $pkgs {
    -uninstall-package $pkg
  }
}`
