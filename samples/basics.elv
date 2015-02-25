# Trivial commands
ls
echo -n 233; echo 333333

# String literals
echo `such\"``literal`
echo "much\n\033[31;1m$cool\033[m"

# Byte pipelines
echo "Albert\nAllan\nAlbraham\nBerlin" | sed `s/l/1/g` | grep e

# Arithmetics
/ 1 0
* (+ 3 4) 6

# Table
put [a b c &key value]
put [a b c &key value][0]
put [a b c &key value][key]

# Variable
var $x string = `SHELL`
println `SUCH `$x`. MUCH COOL`
