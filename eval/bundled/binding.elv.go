package bundled

const bindingElv = `
fn install {
    edit:insert:binding = (edit:binding-table [
        &Default=    $edit:insert:default~
        &F2=         $edit:toggle-quote-paste~
        &Up=         $edit:history:start~
        &Down=       $edit:end-of-history~
        &Right=      $edit:move-dot-right~
        &Left=       $edit:move-dot-left~
        &Home=       $edit:move-dot-sol~
        &Delete=     $edit:kill-rune-right~
        &End=        $edit:move-dot-eol~
        &Tab=        $edit:completion:smart-start~
        &Enter=      $edit:smart-enter~
        &Backspace=  $edit:kill-rune-left~
        &Alt-Up=     $edit:move-dot-up~
        &Alt-Down=   $edit:move-dot-down~
        &Alt-Enter=  $edit:insert-key~
        &Alt-.=      $edit:insert-last-word~
        &Alt-1=      $edit:lastcmd:start~
        &Alt-b=      $edit:move-dot-left-word~
        &Alt-f=      $edit:move-dot-right-word~
        &Ctrl-Right= $edit:move-dot-right-word~
        &Ctrl-Left=  $edit:move-dot-left-word~
        &Ctrl-D=     $edit:return-eof~
        &Ctrl-H=     $edit:kill-rune-left~
        &Ctrl-K=     $edit:kill-line-right~
        &Ctrl-L=     $edit:location:start~
        &Ctrl-N=     $edit:navigation:start~
        &Ctrl-R=     $edit:histlist:start~
        &Ctrl-U=     $edit:kill-line-left~
        &Ctrl-V=     $edit:insert-raw~
        &Ctrl-W=     $edit:kill-word-left~
    ])

    edit:command:binding = (edit:binding-table [
        &Default= $edit:command:default~
        &'$'=     $edit:move-dot-eol~
        &0=       $edit:move-dot-sol~
        &D=       $edit:kill-line-right~
        &b=       $edit:move-dot-left-word~
        &h=       $edit:move-dot-left~
        &i=       $edit:insert:start~
        &j=       $edit:move-dot-down~
        &k=       $edit:move-dot-up~
        &l=       $edit:move-dot-right~
        &w=       $edit:move-dot-right-word~
        &x=       $edit:kill-rune-right~
    ])

    edit:history:binding = (edit:binding-table [
        &Default=  $edit:history:default~
        &Up=       $edit:history:up~
        &Down=     $edit:history:down-or-quit~
        &'Ctrl-['= $edit:insert:start~
    ])

    edit:completion:binding = (edit:binding-table [
        &Default=   $edit:completion:default~
        &Up=        $edit:completion:up~
        &Down=      $edit:completion:down~
        &Right=     $edit:completion:right~
        &Left=      $edit:completion:left~
        &Tab=       $edit:completion:down-cycle~
        &Enter=     $edit:completion:accept~
        &Shift-Tab= $edit:completion:up-cycle~
        &Ctrl-F=    $edit:completion:trigger-filter~
        &'Ctrl-['=  $edit:insert:start~
    ])

    edit:listing:binding = (edit:binding-table [
        &Default=   $edit:listing:default~
        &Up=        $edit:listing:up~
        &Down=      $edit:listing:down~
        &PageUp=    $edit:listing:page-up~
        &PageDown=  $edit:listing:page-down~
        &Tab=       $edit:listing:down-cycle~
        &Enter=     $edit:listing:accept-close~
        &Backspace= $edit:listing:backspace~
        &Shift-Tab= $edit:listing:up-cycle~
        &Alt-Enter= $edit:listing:accept~
        &'Ctrl-['=  $edit:insert:start~
    ])

    edit:histlist:binding = (edit:binding-table [
        &Ctrl-D= $edit:histlist:toggle-dedup~
        &Ctrl-G= $edit:histlist:toggle-case-sensitivity~
    ])

    edit:location:binding = (edit:binding-table [&])

    edit:lastcmd:binding = (edit:binding-table [
        &Alt-1= $edit:lastcmd:accept-line~
    ])

    edit:navigation:binding = (edit:binding-table [
        &Default=   $edit:navigation:default~
        &Up=        $edit:navigation:up~
        &Down=      $edit:navigation:down~
        &Right=     $edit:navigation:right~
        &Left=      $edit:navigation:left~
        &PageUp=    $edit:navigation:page-up~
        &PageDown=  $edit:navigation:page-down~
        &Enter=     $edit:navigation:insert-selected-and-quit~
        &Alt-Up=    $edit:navigation:file-preview-up~
        &Alt-Down=  $edit:navigation:file-preview-down~
        &Alt-Enter= $edit:navigation:insert-selected~
        &Ctrl-F=    $edit:navigation:trigger-filter~
        &Ctrl-H=    $edit:navigation:trigger-shown-hidden~
        &'Ctrl-['=  $edit:insert:start~
    ])

    edit:narrow:binding = (edit:binding-table [&])
}
`

// vi: se ft=elvish et:
