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
# [`from-lines`](#from-lines), although this behavior is subject to change:
#
# ```elvish-transcript
# ~> print "foo\nbar\n" | all
# ▶ foo
# ▶ bar
# ```
#
# @cf one
fn all {|inputs?| }

# Takes exactly one [value input](#value-inputs) and outputs it. If there are
# more than one value inputs, raises an exception.
#
# This function can be used in a similar way to [`all`](#all), but is a better
# choice when you expect that there is exactly one output.
#
# @cf all
fn one {|inputs?| }

# Outputs the first `$n` [value inputs](#value-inputs). If `$n` is larger than
# the number of value inputs, outputs everything.
#
# Examples:
#
# ```elvish-transcript
# ~> range 2 | take 10
# ▶ 0
# ▶ 1
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
# @cf drop
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
# @cf take
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

# Count the number of inputs.
#
# Examples:
#
# ```elvish-transcript
# ~> count lorem # count bytes in a string
# ▶ 5
# ~> count [lorem ipsum]
# ▶ 2
# ~> range 100 | count
# ▶ 100
# ~> seq 100 | count
# ▶ 100
# ```
fn count {|input-list?| }

# Outputs the [value inputs](#value-inputs) sorted in ascending order. The
# sorting process is guaranteed to be
# [stable](https://en.wikipedia.org/wiki/Sorting_algorithm#Stability).
#
# The `&less-than` option, if given, establishes the ordering of the items. Its
# value should be a function that takes two arguments and outputs a single
# boolean indicating whether the first argument is less than the second
# argument. If the function throws an exception, `order` rethrows the exception
# without outputting any value.
#
# If `&less-than` is `$nil` (the default), a builtin comparator equivalent to
# `{|a b| == -1 (compare $a $b) }` is used.
#
# The `&key` option, if given, is a function that gets called with each item to
# be sorted. It must output a single value, which is used for comparison in
# place of the original value. If the function throws an exception, `order`
# rethrows the exception.
#
# Use of `&key` can usually be rewritten to use `&less-than` instead, but using
# `&key` is usually faster because the callback is only called once for each
# element, whereas the `&less-than` callback is called O(n*lg(n)) times on
# average.
#
# If `&key` and `&less-than` are both specified, the output of the `&key`
# callback for each element is passed to the `&less-than` callback.
#
# The `&reverse` option, if true, reverses the order of output.
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
# (The `$"<~"` syntax is a reference to [the `<` function](#num-cmp).)
#
# @cf compare
fn order {|&less-than=$nil &key=$nil &reverse=$false inputs?| }
