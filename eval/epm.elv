-data-dir = ~/.elvish
-repo = $-data-dir/pkgindex
-installed = $-data-dir/epm-installed
-repo-url = https://github.com/elves/pkginfo

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
