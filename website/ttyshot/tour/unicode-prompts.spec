//trim-empty
set edit:rprompt = (constantly ^
    (styled (whoami)"\u00A7"(hostname) inverse))
//prompt
set edit:prompt = {||
    tilde-abbr $pwd
    styled " \u00BB\u00BB " bright-red
}
//wait-for-str »»
//no-enter
# Fancy unicode prompts!
