# Takes [value inputs](#value-inputs), and outputs those values unchanged.
#
# This is an [identity
# function](https://en.wikipedia.org/wiki/Identity_function) for the value
# channel; in other words, `a | all` is equivalent to just `a` if `a` only has
# value output.
#
# This command can be used inside output capture (i.e. `(all)`) to turn value
# inputs into arguments. For example:
#
# ```elvish-transcript
# ~> echo '["foo","bar"] ["lorem","ipsum"]' | from-json
# ▶ [foo bar]
# ▶ [lorem ipsum]
# ~> echo '["foo","bar"] ["lorem","ipsum"]' | from-json | put (all)[0]
# ▶ foo
# ▶ lorem
# ```
#
# The latter pipeline is equivalent to the following:
#
# ```elvish-transcript
# ~> put (echo '["foo","bar"] ["lorem","ipsum"]' | from-json)[0]
# ▶ foo
# ▶ lorem
# ```
#
# In general, when `(all)` appears in the last command of a pipeline, it is
# equivalent to just moving the previous commands of the pipeline into `()`.
# The choice is a stylistic one; the `(all)` variant is longer overall, but can
# be more readable since it allows you to avoid putting an excessively long
# pipeline inside an output capture, and keeps the data flow within the
# pipeline.
#
# Putting the value capture inside `[]` (i.e. `[(all)]`) is useful for storing
# all value inputs in a list for further processing:
#
# ```elvish-transcript
# ~> fn f { var inputs = [(all)]; put $inputs[1] }
# ~> put foo bar baz | f
# ▶ bar
# ```
#
# The `all` command can also take "inputs" from an iterable argument. This can
# be used to flatten lists or strings (although not recursively):
#
# ```elvish-transcript
# ~> all [foo [lorem ipsum]]
# ▶ foo
# ▶ [lorem ipsum]
# ~> all foo
# ▶ f
# ▶ o
# ▶ o
# ```
#
# This can be used together with `(one)` to turn a single iterable value in the
# pipeline into its elements:
#
# ```elvish-transcript
# ~> echo '["foo","bar","lorem"]' | from-json
# ▶ [foo bar lorem]
# ~> echo '["foo","bar","lorem"]' | from-json | all (one)
# ▶ foo
# ▶ bar
# ▶ lorem
# ```
#
# When given byte inputs, the `all` command currently functions like
# [`from-lines`](), although this behavior is subject to change:
#
# ```elvish-transcript
# ~> print "foo\nbar\n" | all
# ▶ foo
# ▶ bar
# ```
#
# See also [`one`]().
fn all {|inputs?| }

# Takes exactly one [value input](#value-inputs) and outputs it. If there are
# more than one value inputs, raises an exception.
#
# This function can be used in a similar way to [`all`](), but is a better
# choice when you expect that there is exactly one output.
#
# See also [`all`]().
fn one {|inputs?| }

# Outputs the first `$n` [value inputs](#value-inputs). If `$n` is larger than
# the number of value inputs, outputs everything.
#
# Examples:
#
# ```elvish-transcript
# ~> range 2 | take 10
# ▶ (num 0)
# ▶ (num 1)
# ~> take 3 [a b c d e]
# ▶ a
# ▶ b
# ▶ c
# ~> use str
# ~> str:split ' ' 'how are you?' | take 1
# ▶ how
# ```
#
# Etymology: Haskell.
#
# See also [`drop`]().
fn take {|n inputs?| }

# Ignores the first `$n` [value inputs](#value-inputs) and outputs the rest.
# If `$n` is larger than the number of value inputs, outputs nothing.
#
# Example:
#
# ```elvish-transcript
# ~> range 10 | drop 8
# ▶ (num 8)
# ▶ (num 9)
# ~> range 2 | drop 10
# ~> drop 2 [a b c d e]
# ▶ c
# ▶ d
# ▶ e
# ~> use str
# ~> str:split ' ' 'how are you?' | drop 1
# ▶ are
# ▶ 'you?'
# ```
#
# Etymology: Haskell.
#
# See also [`take`]().
fn drop {|n inputs?| }

# Replaces consecutive runs of equal values with a single copy. Similar to the
# `uniq` command on Unix.
#
# Examples:
#
# ```elvish-transcript
# ~> put a a b b c | compact
# ▶ a
# ▶ b
# ▶ c
# ~> compact [a a b b c]
# ▶ a
# ▶ b
# ▶ c
# ~> put a b a | compact
# ▶ a
# ▶ b
# ▶ a
# ```
fn compact {|inputs?| }

# Count the number of elements in a container (also known as its _length_), or
# the number of inputs when when argument is given.
#
# Examples:
#
# ```elvish-transcript
# ~> count [lorem ipsum]
# ▶ (num 2)
# ~> count [&foo=bar &lorem=ipsum] # count pairs in a map
# ▶ (num 2)
# ~> count lorem # count bytes in a string
# ▶ (num 5)
# ~> range 100 | count # count inputs
# ▶ (num 100)
# ~> seq 100 | count
# ▶ (num 100)
# ```
fn count {|inputs?| }

# Outputs the [value inputs](#value-inputs) after sorting. The sorting process
# is
# [stable](https://en.wikipedia.org/wiki/Sorting_algorithm#Stability).
#
# By default, `order` sorts the values in ascending order, using the same
# comparator as [`compare`](), which only supports values of the same ordered
# type. Its options modify this behavior:
#
# -   The `&less-than` option, if given, overrides the comparator. Its
#     value should be a function that takes two arguments `$a` and `$b` and
#     outputs a boolean indicating whether `$a` is less than `$b`. If the
#     function throws an exception, `order` rethrows the exception without
#     outputting any value.
#
#     The default behavior of `order` is equivalent to `order &less-than={|a b|
#     == -1 (compare $a $b)}`.
#
# -   The `&total` option, if true, overrides the comparator to be same as
#     `compare &total=$true`, which allows sorting values of mixed types and
#     unordered types. The result groups values by their types. Groups of
#     ordered types are sorted internally, and groups of unordered types retain
#     their original relative order.
#
#     Specifying `&total=$true` is equivalent to specifying `&less-than={|a b|
#     == -1 (compare &total=$true $a $b)}`. It is an error to both specify
#     `&total=$true` and a non-nil `&less-than` callback.
#
# -   The `&key` option, if given, is a function that gets called with each
#     input value. It must output a single value, which is used for comparison
#     in place of the original value. The comparator used can be affected by
#     `$less-than` or `&total`.
#
#     If the function throws an exception, `order` rethrows the exception.
#
#     Use of `&key` can usually be rewritten to use `&less-than` instead, but
#     using `&key` can be faster. The `&key` callback is only called once for
#     each element, whereas the `&less-than` callback is called O(n*lg(n)) times
#     on average.
#
# -   The `&reverse` option, if true, reverses the order of output.
#
# Examples:
#
# ```elvish-transcript
# ~> put foo bar ipsum | order
# ▶ bar
# ▶ foo
# ▶ ipsum
# ~> order [(num 10) (num 1) (num 5)]
# ▶ (num 1)
# ▶ (num 5)
# ▶ (num 10)
# ~> order [[a b] [a] [b b] [a c]]
# ▶ [a]
# ▶ [a b]
# ▶ [a c]
# ▶ [b b]
# ~> order &reverse [a c b]
# ▶ c
# ▶ b
# ▶ a
# ~> put [0 x] [1 a] [2 b] | order &key={|l| put $l[1]}
# ▶ [1 a]
# ▶ [2 b]
# ▶ [0 x]
# ~> order &less-than={|a b| eq $a x } [l x o r x e x m]
# ▶ x
# ▶ x
# ▶ x
# ▶ l
# ▶ o
# ▶ r
# ▶ e
# ▶ m
# ```
#
# Ordering heterogeneous values:
#
# ```elvish-transcript
# //skip-test
# ~> order [a (num 2) c (num 0) b (num 1)]
# Exception: bad value: inputs to "compare" or "order" must be comparable values, but is uncomparable values
#   [tty]:1:1-37: order [a (num 2) c (num 0) b (num 1)]
# ~> order &total [a (num 2) c (num 0) b (num 1)]
# ▶ (num 0)
# ▶ (num 1)
# ▶ (num 2)
# ▶ a
# ▶ b
# ▶ c
# ```
#
# Beware that strings that look like numbers are treated as strings, not
# numbers. To sort strings as numbers, use an explicit `&key` or `&less-than`:
#
# ```elvish-transcript
# ~> order [5 1 10]
# ▶ 1
# ▶ 10
# ▶ 5
# ~> order &key=$num~ [5 1 10]
# ▶ 1
# ▶ 5
# ▶ 10
# ~> order &less-than=$"<~" [5 1 10]
# ▶ 1
# ▶ 5
# ▶ 10
# ```
#
# (The `$"<~"` syntax is a reference to [the `<` function](#num-lt).)
#
# See also [`compare`]().
fn order {|&less-than=$nil &total=$false &key=$nil &reverse=$false inputs?| }

#doc:added-in 0.21
# Calls `$predicate` for each input, outputting those where `$predicate` outputs
# `$true`. Similar to `filter` in some other languages.
#
# The `$predicate` must output a single boolean value.
#
# Examples:
#
# ```elvish-transcript
# ~> use str
# ~> put foo bar foobar | keep-if {|s| str:has-prefix $s f }
# ▶ foo
# ▶ foobar
# ~> keep-if {|s| == 3 (count $s) } [foo bar foobar]
# ▶ foo
# ▶ bar
# ```
fn keep-if {|predicate inputs?| }
