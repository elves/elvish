# A sample file to test syntax highlighting.

nop "double \n quoted" and 'single '' quoted' # comment

# Various variable contexts
nop $pid
var var-name = { var fn-name~ = {var not-var-name} }
nop (set var-name = foo | tmp var-name = bar); del var-name
for var-name [] { }
try { } catch var-name { }

# Builtin functions
!= a (nop b) | echo c

# Builtin special command
and a b # "operator"
use re # "other"
if a { } elif b { } else { }
try { } except err { } else { } finally { }

# Metacharacters
echo ** () []

# Regression tests
set-env # should highlight entire set-env
set-foo # should highlight nothing
