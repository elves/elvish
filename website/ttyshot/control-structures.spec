if $true { echo good } else { echo bad }
#prompt
for x [lorem ipsum] {
  echo $x.pdf
}
#prompt
try {
  fail 'bad error'
} catch e {
  echo error $e
} else {
  echo ok
}
