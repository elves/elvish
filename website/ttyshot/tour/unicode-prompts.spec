set edit:rprompt = (constantly ^
  (styled (whoami)✸(hostname) inverse))
#prompt
set edit:prompt = {
  tilde-abbr $pwd
  styled '❱ ' bright-red
}
#prompt
#no-enter
# Fancy unicode prompts!
