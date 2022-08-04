# This is the interactive configuration for generating Elvish "ttyshots".

{

use store
# Populate the interactive location history.
range 9 | each {|_| store:add-dir $E:HOME }
range 9 | each {|_| store:add-dir $E:HOME/elvish }
range 8 | each {|_| store:add-dir /tmp }
range 7 | each {|_| store:add-dir $E:HOME/.config/elvish }
range 6 | each {|_| store:add-dir $E:HOME/.local/share/elvish }
range 5 | each {|_| store:add-dir /usr }
range 4 | each {|_| store:add-dir /usr/local/bin }
range 3 | each {|_| store:add-dir /usr/local/share }
range 2 | each {|_| store:add-dir /usr/local }
range 1 | each {|_| store:add-dir /opt }

# Populate the interactive command history.
set @_ = (range 5 | each {|_|
    store:add-cmd 'randint 1 10'
    store:add-cmd 'echo (styled warning: red) bumpy road'
    store:add-cmd 'echo "hello\nbye" > /tmp/x'
    store:add-cmd 'from-lines < /tmp/x'
    store:add-cmd 'cd /tmp'
    store:add-cmd 'cd ~/elvish'
    store:add-cmd 'git branch'
    store:add-cmd 'git checkout .'
    store:add-cmd 'git commit'
    store:add-cmd 'git status'
    store:add-cmd 'git status'
    store:add-cmd 'git status'
    store:add-cmd 'git status'
    store:add-cmd 'git status'
    store:add-cmd 'git status'
    store:add-cmd 'git status'
    store:add-cmd 'git status'
    store:add-cmd 'git status'
    store:add-cmd 'cd /usr/local/bin'
    store:add-cmd 'echo $pwd'
    store:add-cmd '* (+ 3 4) (- 100 94)'
    store:add-cmd 'make'
    store:add-cmd 'make'
    store:add-cmd 'make'
    store:add-cmd 'make'
    store:add-cmd 'make'
    store:add-cmd 'make'
    store:add-cmd 'make'
    store:add-cmd 'make'
    store:add-cmd 'make'
    store:add-cmd 'math:min 3 1 30'
})

} # use store

# Sync the history we just manufactured with this elvish process.
edit:history:fast-forward

set edit:global-binding[Alt-q] = {
    tmux capture-pane -epN > ~/.tmp/ttyshot.raw
    exit
}

set edit:max-height = 16
set edit:navigation:width-ratio = [8 18 30]

set edit:before-readline = [$@edit:before-readline {
    echo '[PROMPT]'
}]

set edit:rprompt = (constantly (styled 'elf@host' inverse))

# These functions are used in some of the ttyshot scripts to ensure consistent
# output that doesn't leak info about the machine used to create the ttyshot.
fn whoami { echo elf }
fn hostname { echo host.example.com }
