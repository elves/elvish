curl -sL api.github.com/repos/elves/elvish/issues |
  all (from-json) |
  each {|x| echo (exact-num $x[number]): $x[title] } |
  head -n 7
//prompt
