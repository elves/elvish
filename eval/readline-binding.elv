fn bind-mode [m k f]{
    edit:binding[$m][$k] = $f
}

fn bind [k f]{
    bind-mode insert $k $f
}

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
bind Alt-n $edit:nav:&start
bind Alt-l $edit:loc:&start

bind-mode completion Ctrl-B $edit:compl:&left
bind-mode completion Ctrl-F $edit:compl:&right
bind-mode completion Ctrl-N $edit:compl:&down
bind-mode completion Ctrl-P $edit:compl:&up
bind-mode completion Alt-f  $edit:compl:&trigger-filter

bind-mode navigation Ctrl-B $edit:nav:&left
bind-mode navigation Ctrl-F $edit:nav:&right
bind-mode navigation Ctrl-N $edit:nav:&down
bind-mode navigation Ctrl-P $edit:nav:&up
bind-mode navigation Alt-f  $edit:nav:&trigger-filter

bind-mode history Ctrl-N $edit:history:&down-or-quit
bind-mode history Ctrl-P $edit:history:&up
bind-mode history Ctrl-G $edit:insert:&start

# Binding for the listing "super mode".
bind-mode listing Ctrl-N $edit:listing:&down
bind-mode listing Ctrl-P $edit:listing:&up
bind-mode listing Ctrl-V $edit:listing:&page-down
bind-mode listing Alt-v  $edit:listing:&page-up
bind-mode listing Ctrl-G $edit:insert:&start

bind-mode histlist Alt-g $edit:histlist:&toggle-case-sensitivity
bind-mode histlist Alt-d $edit:histlist:&toggle-dedup
