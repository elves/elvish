# Implementation of location, history and lastcmd mode using the new
# -narrow-read mode. One advantage of this is that it allows the
# execution of arbitrary hooks before or after each mode.
#
# Usage:
#   use narrow
#   narrow:bind_keys
#
# narrow:bind_keys binds keys for location, history and lastcmd
# modes. Without options, it uses the default bindings (same as the
# default bindings for edit:location, edit:history and edit:lastcmd),
# but different keys can be specified with the options. To disable a
# binding, specify its key as "".
# Example:
#   narrow:bind_keys &location=Alt-l &lastcmd="" 

# Hooks
# Each hook variable is an array which must contain lambdas, all of
# which will be executed in sequence before and after the
# corresponding mode.
# Example (list the new directory after switching to it in location mode):
#    narrow:after-location = [ $@narrow:after-location { edit:insert-at-dot "ls"; edit:smart-enter } ]
before-location = []
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


fn -bind_i [k f]{
  edit:insert:binding[$k] = $f
}

fn -bind_n [k f]{
  edit:narrow:binding[$k] = $f
}

# Bind keys for location, history and lastcmd modes. Without
# options, it uses the default bindings, but different keys
# can be specified with the options. To disable a binding,
# specify its key as "".
# Example:
#   narrow:bind_keys &location=Alt-l &lastcmd=""
fn bind_keys [&location=C-l &history=C-r &lastcmd=M-1]{
  if (> (count $location) 0) { -bind_i $location narrow:location }
  if (> (count $history) 0)  { -bind_i $history  narrow:history }
  if (> (count $lastcmd) 0)  { -bind_i $lastcmd  narrow:lastcmd }
}

# Set up some default useful bindings for narrow mode
-bind_n Up        $edit:narrow:&up
-bind_n PageUp    $edit:narrow:&page-up
-bind_n Down      $edit:narrow:&down
-bind_n PageDown  $edit:narrow:&page-down
-bind_n Tab       $edit:narrow:&down-cycle
-bind_n S-Tab     $edit:narrow:&up-cycle
-bind_n Backspace $edit:narrow:&backspace
-bind_n Enter     $edit:narrow:&accept-close
-bind_n M-Enter   $edit:narrow:&accept
-bind_n default   $edit:narrow:&default
-bind_n "C-["     $edit:insert:&start
-bind_n C-G       $edit:narrow:&toggle-ignore-case
-bind_n C-D       $edit:narrow:&toggle-ignore-duplication
