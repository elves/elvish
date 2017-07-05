fn location {
    candidates = [(dir-history | each {
	score = (splits $0[score] &sep=. | take 1)
        put [
            &content=$0[path]
            &display=$score" "$0[path]
	]
    })]

    edit:-narrow-read {
        put $@candidates
    } {
        cd $0[content]
    } &modeline="Location "
}

fn history {
    candidates = [(edit:command-history | each {
        put [
	    &content=$0[cmd]
	    &display=$0[id]" "$0[cmd]
        ]
    })]

    edit:-narrow-read {
        put $@candidates
    } {
        edit:replace-input $0[content]
    } &modeline="History " &keep-bottom=$true
}

fn lastcmd {
    last = (edit:command-history -1)
    cmd = [
            &content=$last[cmd]
            &display="-1 "$last[cmd]
	    &filter-text=""
        ]
    index = 0
    candidates = [$cmd ( edit:wordify $last[cmd] | each {
	put [
            &content=$0
            &display=$index" "$0
            &filter-text=$index
	]
	index = (+ $index 1)
    })]
    edit:-narrow-read {
        put $@candidates
    } {
        edit:replace-input $0[content]
    } &modeline="Lastcmd " &auto-commit=$true &bindings=[&M-1={ edit:narrow:accept-close }]
}


# TODO: separate bindings from functions

fn bind [m k f]{
    edit:binding[$m][$k] = $f
}

bind insert C-l       narrow:location
bind insert C-r       narrow:history
bind insert M-1       narrow:lastcmd

bind narrow Up        $edit:narrow:&up
bind narrow PageUp    $edit:narrow:&page-up
bind narrow Down      $edit:narrow:&down
bind narrow PageDown  $edit:narrow:&page-down
bind narrow Tab       $edit:narrow:&down-cycle
bind narrow S-Tab     $edit:narrow:&up-cycle
bind narrow Backspace $edit:narrow:&backspace
bind narrow Enter     $edit:narrow:&accept-close
bind narrow M-Enter   $edit:narrow:&accept
bind narrow default   $edit:narrow:&default
bind narrow "C-["     $edit:insert:&start
bind narrow C-G       $edit:narrow:&toggle-ignore-case
bind narrow C-D       $edit:narrow:&toggle-ignore-duplication
