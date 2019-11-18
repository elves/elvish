package cliedit

// Elvish code for default bindings, assuming the editor ns as the global ns.
const defaultBindingsElv = `
insert:binding = (binding-table [
  &Left=  $move-dot-left~
  &Right= $move-dot-right~

  &Ctrl-Left=  $move-dot-left-word~
  &Ctrl-Right= $move-dot-right-word~
  &Alt-Left=   $move-dot-left-word~
  &Alt-Right=  $move-dot-right-word~
  &Alt-b=      $move-dot-left-word~
  &Alt-f=      $move-dot-right-word~

  &Home= $move-dot-sol~
  &End=  $move-dot-eol~

  &Backspace= $kill-rune-left~
  &Delete=    $kill-rune-right~
  &Ctrl-W=    $kill-word-left~
  &Ctrl-U=    $kill-line-left~
  &Ctrl-K=    $kill-line-right~

  &Ctrl-V= $insert-raw~

  &Alt-,=  $lastcmd:start~
  &Ctrl-R= $histlist:start~
  &Ctrl-L= $location:start~
  &Ctrl-N= $navigation:start~
  &Tab=    $completion:start~
  &Up=     $history:start~

  &Enter=   $smart-enter~
  &Ctrl-D=  $return-eof~
])

listing:binding = (binding-table [
  &Up=        $listing:up~
  &Down=      $listing:down~
  &Tab=       $listing:down-cycle~
  &Shift-Tab= $listing:up-cycle~
  &Ctrl-'['=  $close-listing~
])

histlist:binding = (binding-table [
  &Ctrl-D= $histlist:toggle-dedup~
])

navigation:binding = (binding-table [
  &Ctrl-'['= $close-listing~
])

completion:binding = (binding-table [
  &Ctrl-'['= $completion:close~
  &Down=     $completion:down~
  &Up=       $completion:up~
  &Tab=      $completion:down-cycle~
  &Shift-Tab=$completion:up-cycle~
  &Left=     $completion:left~
  &Right=    $completion:right~
])

history:binding = (binding-table [
  &Up=       $history:up~
  &Down=     $history:down~
  &Enter=    $history:accept~
  &Ctrl-'['= $history:close~
])

#  &Up=        $listing:up~
#  &Down=      $listing:down~
#  &Tab=       $listing:down-cycle~
#  &Shift-Tab= $listing:up-cycle~
#
#  &Ctrl-F=    $listing:toggle-filtering~
#
#  &Alt-Enter= $listing:accept~
#  &Enter=     $listing:accept-close~
#  &Alt-,=     $listing:accept-close~
#  &Ctrl-'['=  $reset-mode~
#
#  &Default=   $listing:default~
`

// vi: set et:
