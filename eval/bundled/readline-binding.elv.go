package bundled

const readlineBindingElv = `
b=[k f]{ edit:insert:binding[$k] = $f } {
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
`
