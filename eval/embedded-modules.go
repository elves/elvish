package eval

var embeddedModules = map[string]string{
	"embedded:acme": `fn acme {
    echo 'So this'
    put works.
}
`,
	"embedded:readline-binding": `fn bind-mode [m k f]{
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
bind-mode history Ctrl-G $le:insert:&start

bind-mode histlist Ctrl-N $le:histlist:&down
bind-mode histlist Ctrl-P $le:histlist:&up
bind-mode histlist Ctrl-G $le:insert:&start
bind-mode histlist Alt-g $le:histlist:&toggle-case-sensitivity
bind-mode histlist Alt-d $le:histlist:&toggle-dedup
bind-mode histlist Ctrl-V $le:histlist:&page-down
bind-mode histlist Alt-v $le:histlist:&page-up

bind-mode loc Ctrl-N $le:loc:&down
bind-mode loc Ctrl-P $le:loc:&up
bind-mode loc Ctrl-V $le:loc:&page-down
bind-mode loc Alt-v $le:loc:&page-up
bind-mode loc Ctrl-G $le:insert:&start

bind-mode bang Ctrl-N $le:bang:&down
bind-mode bang Ctrl-P $le:bang:&up
bind-mode bang Ctrl-V $le:bang:&page-down
bind-mode bang Alt-v $le:bang:&page-up
bind-mode bang Ctrl-G $le:insert:&start
`,
}
