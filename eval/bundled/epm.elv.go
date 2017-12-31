package bundled

const epmElv = `
-data-dir = ~/.elvish
-installed = $-data-dir/epm-installed

fn -info [text]{
    print (edit:styled '=> ' green)
    echo $text
}

fn -error [text]{
    print (edit:styled '=> ' red)
    echo $text
}

fn add-installed [@pkgs]{
    if (eq $pkgs []) {
        -error 'Must specify at least one package.'
        return
    }
    for pkg $pkgs {
        echo $pkg >> $-installed
    }
}

fn -get-url [pkg]{
    put https://$pkg
}

fn -install-one [pkg]{
    dest = $-data-dir/lib/$pkg
    if ?(test -e $dest) {
        -error 'Package '$pkg' already exists locally.'
        return
    }
    -info 'Installing '$pkg
    mkdir -p $dest
    git clone (-get-url $pkg) $dest
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
`
