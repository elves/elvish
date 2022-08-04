if $true { echo good } else { echo bad }
#prompt
for x [lorem ipsum] { put $x.pdf }
#prompt
try {
  fail 'bad error'
} catch e {
  put $e
} finally {
  put done
}
