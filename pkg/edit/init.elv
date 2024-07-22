set after-command = [
  # Capture the most recent interactive command duration in $edit:command-duration
  # as a convenience for prompt functions. Note: The first time this is run is after
  # shell.sourceRC() finishes so the initial value of command-duration is the time
  # to execute the user's interactive configuration script.
  {|m|
    set command-duration = $m[duration]
  }
]

set completion:arg-completer = [
  &cd=         $complete-dirname~
  &sudo=       $complete-sudo~
  &doc:show=   {|@a| use doc; doc:-symbols }
  &doc:source= {|@a| use doc; doc:-symbols }
]

set global-binding = (binding-table [
  &Ctrl-'['= $close-mode~
  &Alt-x=    $minibuf:start~
])

set insert:binding = (binding-table [
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
  &Ctrl-H=    $kill-rune-left~
  &Delete=    $kill-rune-right~
  &Ctrl-W=    $kill-word-left~
  &Ctrl-U=    $kill-line-left~
  &Ctrl-K=    $kill-line-right~

  &Ctrl-V= $insert-raw~

  &Alt-,=  $lastcmd:start~
  &Alt-.=  $insert-last-word~
  &Ctrl-R= $histlist:start~
  &Ctrl-L= $location:start~
  &Ctrl-N= $navigation:start~
  &Tab=    $completion:smart-start~
  &Up=     $history:start~
  &Down=   $end-of-history~

  &Alt-Enter={ insert-at-dot "\n" }

  &Ctrl-A= $apply-autofix~

  &Enter=   $smart-enter~
  &Ctrl-D=  $return-eof~
])

set command:binding = (binding-table [
  &'$'= $move-dot-eol~
  &0=   $move-dot-sol~
  &D=   $kill-line-right~
  &b=   $move-dot-left-word~
  &h=   $move-dot-left~
  &i=   $close-mode~
  &a=   { $move-dot-right~; $close-mode~ }
  &j=   $move-dot-down~
  &k=   $move-dot-up~
  &l=   $move-dot-right~
  &w=   $move-dot-right-word~
  &x=   $kill-rune-right~
])

set listing:binding = (binding-table [
  &Up=        $listing:up~
  &Down=      $listing:down~
  &PageUp=    $listing:page-up~
  &PageDown=  $listing:page-down~
  &Tab=       $listing:down-cycle~
  &Shift-Tab= $listing:up-cycle~
])

set histlist:binding = (binding-table [
  &Ctrl-D= $histlist:toggle-dedup~
])

set navigation:binding = (binding-table [
  &Left=     $navigation:left~
  &Right=    $navigation:right~
  &Up=       $navigation:up~
  &Down=     $navigation:down~
  &PageUp=   $navigation:page-up~
  &PageDown= $navigation:page-down~
  &Alt-Up=   $navigation:file-preview-up~
  &Alt-Down= $navigation:file-preview-down~
  &Enter=    $navigation:insert-selected-and-quit~
  &Alt-Enter= $navigation:insert-selected~
  &Ctrl-F=   $navigation:trigger-filter~
  &Ctrl-H=   $navigation:trigger-shown-hidden~
])

set completion:binding = (binding-table [
  &Down=     $completion:down~
  &Up=       $completion:up~
  &Tab=      $completion:down-cycle~
  &Shift-Tab=$completion:up-cycle~
  &Left=     $completion:left~
  &Right=    $completion:right~
])

set history:binding = (binding-table [
  &Up=       $history:up~
  &Down=     $history:down-or-quit~
  &Ctrl-'['= $close-mode~
])

set lastcmd:binding = (binding-table [
  &Alt-,=  $listing:accept~
])

set -instant:binding = (binding-table [
  &
])

# TODO: Avoid duplicating the bindings here by having a base binding table
# shared by insert and minibuf modes (like how the listing modes all share
# listing:binding).
set minibuf:binding = (binding-table [
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
  &Ctrl-H=    $kill-rune-left~
  &Delete=    $kill-rune-right~
  &Ctrl-W=    $kill-word-left~
  &Ctrl-U=    $kill-line-left~
  &Ctrl-K=    $kill-line-right~

  &Ctrl-V= $insert-raw~

  &Alt-,=  $lastcmd:start~
  &Alt-.=  $insert-last-word~
  &Ctrl-R= $histlist:start~
  &Ctrl-L= $location:start~
  &Ctrl-N= $navigation:start~
  &Tab=    $completion:smart-start~
  &Up=     $history:start~
])
