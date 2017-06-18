fn bind-mode [m k f]{
    le:binding[$m][$k] = $f
}

fn bind [k f]{
    bind-mode insert $k $f
}

bind Ctrl-A $le:&move-dot-sol
bind Ctrl-B $le:&move-dot-left
bind Ctrl-D {
    if (> (count $le:current-command) 0) {
        le:kill-rune-right
    } else {
        le:return-eof
    }
}
bind Ctrl-E $le:&move-dot-eol
bind Ctrl-F $le:&move-dot-right
bind Ctrl-H $le:&kill-rune-left
bind Ctrl-L { clear > /dev/tty }
bind Ctrl-N $le:&end-of-history
# TODO: ^O
bind Ctrl-P $le:history:&start
# TODO: ^S ^T ^X family ^Y ^_
bind Alt-b  $le:&move-dot-left-word
# TODO Alt-c Alt-d
bind Alt-f  $le:&move-dot-right-word
# TODO Alt-l Alt-r Alt-u

# Ctrl-N and Ctrl-L occupied by readline binding, bind to Alt- instead.
bind Alt-n $le:nav:&start
bind Alt-l $le:loc:&start

bind-mode completion Ctrl-B $le:compl:&left
bind-mode completion Ctrl-F $le:compl:&right
bind-mode completion Ctrl-N $le:compl:&down
bind-mode completion Ctrl-P $le:compl:&up
bind-mode completion Alt-f  $le:compl:&trigger-filter

bind-mode navigation Ctrl-B $le:nav:&left
bind-mode navigation Ctrl-F $le:nav:&right
bind-mode navigation Ctrl-N $le:nav:&down
bind-mode navigation Ctrl-P $le:nav:&up
bind-mode navigation Alt-f  $le:nav:&trigger-filter

bind-mode history Ctrl-N $le:history:&down-or-quit
bind-mode history Ctrl-P $le:history:&up
