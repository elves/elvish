binding = [&]

fn bind [k f]{
    binding[$k] = $f
}

binding=$edit:insert:binding {
    bind Ctrl-A $edit:&move-dot-sol
    bind Ctrl-B $edit:&move-dot-left
    bind Ctrl-D {
        if (> (count $edit:current-command) 0) {
            edit:kill-rune-right
        } else {
            edit:return-eof
        }
    }
    bind Ctrl-E $edit:&move-dot-eol
    bind Ctrl-F $edit:&move-dot-right
    bind Ctrl-H $edit:&kill-rune-left
    bind Ctrl-L { clear > /dev/tty }
    bind Ctrl-N $edit:&end-of-history
    # TODO: ^O
    bind Ctrl-P $edit:history:&start
    # TODO: ^S ^T ^X family ^Y ^_
    bind Alt-b  $edit:&move-dot-left-word
    # TODO Alt-c Alt-d
    bind Alt-f  $edit:&move-dot-right-word
    # TODO Alt-l Alt-r Alt-u

    # Ctrl-N and Ctrl-L occupied by readline binding, bind to Alt- instead.
    bind Alt-n $edit:navigation:&start
    bind Alt-l $edit:location:&start
}

binding=$edit:completion:binding {
    bind Ctrl-B $edit:completion:&left
    bind Ctrl-F $edit:completion:&right
    bind Ctrl-N $edit:completion:&down
    bind Ctrl-P $edit:completion:&up
    bind Alt-f  $edit:completion:&trigger-filter
}

binding=$edit:navigation:binding {
    bind Ctrl-B $edit:navigation:&left
    bind Ctrl-F $edit:navigation:&right
    bind Ctrl-N $edit:navigation:&down
    bind Ctrl-P $edit:navigation:&up
    bind Alt-f  $edit:navigation:&trigger-filter
}

binding=$edit:history:binding {
    bind Ctrl-N $edit:history:&down-or-quit
    bind Ctrl-P $edit:history:&up
    bind Ctrl-G $edit:insert:&start
}

# Binding for the listing "super mode".
binding=$edit:listing:binding {
    bind Ctrl-N $edit:listing:&down
    bind Ctrl-P $edit:listing:&up
    bind Ctrl-V $edit:listing:&page-down
    bind Alt-v  $edit:listing:&page-up
    bind Ctrl-G $edit:insert:&start
}

binding=$edit:histlist:binding {
    bind Alt-g $edit:histlist:&toggle-case-sensitivity
    bind Alt-d $edit:histlist:&toggle-dedup
}
