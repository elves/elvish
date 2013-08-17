Metacharacter viability in shells
=================================

```
~ +5
    sh: shorthand for $HOME at beginning of a word (+5)
` +5
    sh: command substitution (+5)
! +2
    bash, zsh: history expansion (+3)
    plan9: address delimiter (-1)
@ -2
    ssh: username@host (-2)
# +5
    sh: comment (+5)
$ +5
    sh: variable leader (+5)
% +1
    fish: job leader (+1)
^ +7
    bash: history substitution (+4)
    fish: stderr redirec (+1)
    rc: word joiner (+2)
& +5
    sh: background job (+5)
* +5
    sh: glob
() +1
    zsh: glob (+1)
    fish: command substitution (+1)
    common in natural languages (-1)
- -5
    used in --option (-5)
_ -2
    used in file_names (-2)
+ -2
    less, sh: +option (-2)
= -5
    used in --option-key=value (-5)
{} +4
    bash: brace expansion (+4)
[] +4
    sh: glob (+5)
    ipv6 address (-1)
| +5
    sh: pipeline
\ +5
    sh: escape
: -3
    lsof: port (-2)
    ipv6 address (-1)
; +5
    sh: command terminator (+5)
" +5
    sh: string (+5)
' +5
    sh: string (+5)
, -3
    --option-key=v1,v2
. -5
    a.out
/ -5
    file path
<> +5
    sh: redir (+5)
? +5
    sh: glob
```
