package bundled

const narrowElv = `
# Implementation of location, history and lastcmd mode using the new
# -narrow-read mode. One advantage of this is that it allows the
# execution of arbitrary hooks before or after each mode.
#
# Usage:
#   use narrow
#   narrow:bind-trigger-keys
#
# narrow:bind-trigger-keys binds keys for location, history and lastcmd
# modes. Without options, it uses the default bindings (same as the
# default bindings for edit:location, edit:history and edit:lastcmd),
# but different keys can be specified with the options. To disable a
# binding, specify its key as "".
# Example:
#   narrow:bind-trigger-keys &location=Alt-l &lastcmd=""

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
          &display=$score" "(tilde-abbr $arg[path])
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
	        &display=$arg[id]" "$arg[cmd]
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

# Bind keys for location, history and lastcmd modes. Without
# options, it uses the default bindings, but different keys
# can be specified with the options. To disable a binding,
# specify its key as "".
# Example:
#   narrow:bind-trigger-keys &location=Alt-l &lastcmd=""
fn bind-trigger-keys [&location=C-l &history=C-r &lastcmd=M-1]{
  if (not-eq $location "") { -bind-insert $location $location~ }
  if (not-eq $history "")  { -bind-insert $history  $history~ }
  if (not-eq $lastcmd "")  { -bind-insert $lastcmd  $lastcmd~ }
}

# Set up some default useful bindings for narrow mode
-bind Up        $edit:narrow:up~
-bind PageUp    $edit:narrow:page-up~
-bind Down      $edit:narrow:down~
-bind PageDown  $edit:narrow:page-down~
-bind Tab       $edit:narrow:down-cycle~
-bind S-Tab     $edit:narrow:up-cycle~
-bind Backspace $edit:narrow:backspace~
-bind Enter     $edit:narrow:accept-close~
-bind M-Enter   $edit:narrow:accept~
-bind Default   $edit:narrow:default~
-bind "C-["     $edit:insert:start~
-bind C-G       $edit:narrow:toggle-ignore-case~
-bind C-D       $edit:narrow:toggle-ignore-duplication~
`
