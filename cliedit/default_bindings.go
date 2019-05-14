package cliedit

// Elvish code for default bindings, assuming the editor ns as the global ns.
const defaultBindingsElv = `
insert:binding = (binding-map [
  &Left=  $move-left~
  &Right= $move-right~

  &Ctrl-Left=  $move-left-word~
  &Ctrl-Right= $move-right-word~
  &Alt-Left=   $move-left-word~
  &Alt-Right=  $move-right-word~
  &Alt-b=      $move-left-word~
  &Alt-f=      $move-right-word~

  &Home= $move-sol~
  &End=  $move-eol~

  &Backspace= $kill-left~
  &Delete=    $kill-right~
  &Ctrl-W=    $kill-left-word~
  &Ctrl-U=    $kill-sol~
  &Ctrl-K=    $kill-eol~

  &Alt-,=  $lastcmd:start~
  &Ctrl-R= $histlist:start~

  &Ctrl-D=  $commit-eof~
  &Default= $insert:default-handler~
])

listing:binding = (binding-map [
  &Up=        $listing:up~
  &Down=      $listing:down~
  &Tab=       $listing:down-cycle~
  &Shift-Tab= $listing:up-cycle~

  &Ctrl-F=    $listing:toggle-filtering~

  &Alt-Enter= $listing:accept~
  &Enter=     $listing:accept-close~
  &Alt-,=     $listing:accept-close~
  &Ctrl-'['=  $reset-mode~

  &Default=   $listing:default~
])
`

// vi: set et:
