# Constructs a namespace from `$map`, using the keys as variable names and the
# values as their values. Examples:
#
# ```elvish-transcript
# ~> var n = (ns [&name=value])
# ~> put $n[name]
# ▶ value
# ~> var n: = (ns [&name=value])
# ~> put $n:name
# ▶ value
# ```
fn ns {|map| }

# Outputs a map from the [value inputs](#value-inputs), each of which must be
# an iterable value with with two elements. The first element of each value
# is used as the key, and the second element is used as the value.
#
# If the same key appears multiple times, the last value is used.
#
# Examples:
#
# ```elvish-transcript
# ~> make-map [[k v]]
# ▶ [&k=v]
# ~> make-map [[k v1] [k v2]]
# ▶ [&k=v2]
# ~> put [k1 v1] [k2 v2] | make-map
# ▶ [&k1=v1 &k2=v2]
# ~> put aA bB | make-map
# ▶ [&a=A &b=B]
# ```
fn make-map {|input?| }

# Outputs a list created from adding values in `$more` to the end of `$list`.
#
# The output is the same as `[$@list $more...]`, but the time complexity is
# guaranteed to be O(m), where m is the number of values in `$more`.
#
# Examples:
#
# ```elvish-transcript
# ~> conj [] a
# ▶ [a]
# ~> conj [a b] c d
# ▶ [a b c d]
# ```
#
# Etymology: [Clojure](https://clojuredocs.org/clojure.core/conj).
fn conj {|list @more| }

# Output a slightly modified version of `$container`, such that its value at `$k`
# is `$v`. Applies to both lists and to maps.
#
# When `$container` is a list, `$k` may be a negative index. However, slice is not
# yet supported.
#
# ```elvish-transcript
# ~> assoc [foo bar quux] 0 lorem
# ▶ [lorem bar quux]
# ~> assoc [foo bar quux] -1 ipsum
# ▶ [foo bar ipsum]
# ~> assoc [&k=v] k v2
# ▶ [&k=v2]
# ~> assoc [&k=v] k2 v2
# ▶ [&k=v &k2=v2]
# ```
#
# Etymology: [Clojure](https://clojuredocs.org/clojure.core/assoc).
#
# See also [`dissoc`]().
fn assoc {|container k v| }

# Output a slightly modified version of `$map`, with the key `$k` removed. If
# `$map` does not contain `$k` as a key, the same map is returned.
#
# ```elvish-transcript
# ~> dissoc [&foo=bar &lorem=ipsum] foo
# ▶ [&lorem=ipsum]
# ~> dissoc [&foo=bar &lorem=ipsum] k
# ▶ [&foo=bar &lorem=ipsum]
# ```
#
# See also [`assoc`]().
fn dissoc {|map k| }

# Determine whether `$value` is a value in `$container`.
#
# Examples, maps:
#
# ```elvish-transcript
# ~> has-value [&k1=v1 &k2=v2] v1
# ▶ $true
# ~> has-value [&k1=v1 &k2=v2] k1
# ▶ $false
# ```
#
# Examples, lists:
#
# ```elvish-transcript
# ~> has-value [v1 v2] v1
# ▶ $true
# ~> has-value [v1 v2] k1
# ▶ $false
# ```
#
# Examples, strings:
#
# ```elvish-transcript
# ~> has-value ab b
# ▶ $true
# ~> has-value ab c
# ▶ $false
# ```
fn has-value {|container value| }

# Determine whether `$key` is a key in `$container`. A key could be a map key or
# an index on a list or string. This includes a range of indexes.
#
# Examples, maps:
#
# ```elvish-transcript
# ~> has-key [&k1=v1 &k2=v2] k1
# ▶ $true
# ~> has-key [&k1=v1 &k2=v2] v1
# ▶ $false
# ```
#
# Examples, lists:
#
# ```elvish-transcript
# ~> has-key [v1 v2] 0
# ▶ $true
# ~> has-key [v1 v2] 1
# ▶ $true
# ~> has-key [v1 v2] 2
# ▶ $false
# ~> has-key [v1 v2] 0..2
# ▶ $true
# ~> has-key [v1 v2] 0..3
# ▶ $false
# ```
#
# Examples, strings:
#
# ```elvish-transcript
# ~> has-key ab 0
# ▶ $true
# ~> has-key ab 1
# ▶ $true
# ~> has-key ab 2
# ▶ $false
# ~> has-key ab 0..2
# ▶ $true
# ~> has-key ab 0..3
# ▶ $false
# ```
fn has-key {|container key| }

#//skip-test
# Put all keys of `$map` on the structured stdout.
#
# Example:
#
# ```elvish-transcript
# ~> keys [&a=foo &b=bar &c=baz]
# ▶ a
# ▶ c
# ▶ b
# ```
#
# Note that there is no guaranteed order for the keys of a map.
fn keys {|map| }
