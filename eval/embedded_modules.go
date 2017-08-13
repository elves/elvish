package eval

var embeddedModules = map[string]string{
	"narrow": `before-location = []
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

fn bind-trigger-keys [&location=C-l &history=C-r &lastcmd=M-1]{
  if (not-eq $location "") { -bind-insert $location narrow:location }
  if (not-eq $history "")  { -bind-insert $history  narrow:history }
  if (not-eq $lastcmd "")  { -bind-insert $lastcmd  narrow:lastcmd }
}

-bind Up        $edit:narrow:&up
-bind PageUp    $edit:narrow:&page-up
-bind Down      $edit:narrow:&down
-bind PageDown  $edit:narrow:&page-down
-bind Tab       $edit:narrow:&down-cycle
-bind S-Tab     $edit:narrow:&up-cycle
-bind Backspace $edit:narrow:&backspace
-bind Enter     $edit:narrow:&accept-close
-bind M-Enter   $edit:narrow:&accept
-bind default   $edit:narrow:&default
-bind "C-["     $edit:insert:&start
-bind C-G       $edit:narrow:&toggle-ignore-case
-bind C-D       $edit:narrow:&toggle-ignore-duplication
`,
	"readline-binding": `binding = [&]

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
`,
}
