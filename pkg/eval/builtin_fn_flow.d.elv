# Run several callables in parallel, and wait for all of them to finish.
#
# If one or more callables throw exceptions, the other callables continue running,
# and a composite exception is thrown when all callables finish execution.
#
# The behavior of `run-parallel` is consistent with the behavior of pipelines,
# except that it does not perform any redirections.
#
# Here is an example that lets you pipe the stdout and stderr of a command to two
# different commands in order to independently capture the output of each byte stream:
#
# ```elvish-transcript
# ~> use file
# ~> fn capture {|f|
#      var pout = (file:pipe)
#      var perr = (file:pipe)
#      var out err
#      run-parallel {
#        $f > $pout[w] 2> $perr[w]
#        file:close $pout[w]
#        file:close $perr[w]
#      } {
#        set out = (slurp < $pout[r])
#        file:close $pout[r]
#      } {
#        set err = (slurp < $perr[r])
#        file:close $perr[r]
#      }
#      put $out $err
#    }
# ~> capture { echo stdout-test; echo stderr-test >&2 }
# ▶ "stdout-test\n"
# ▶ "stderr-test\n"
# ```
#
# This command is intended for doing a fixed number of heterogeneous things in
# parallel. If you need homogeneous parallel processing of possibly unbound data,
# use `peach` instead.
#
# See also [`peach`]().
fn run-parallel {|@callable| }

# Calls `$f` on each [value input](#value-inputs).
#
# An exception raised from [`break`]() is caught by `each`, and will cause it to
# terminate early.
#
# An exception raised from [`continue`]() is swallowed and can be used to
# terminate a single iteration early.
#
# Examples:
#
# ```elvish-transcript
# ~> range 5 8 | each {|x| * $x $x }
# ▶ (num 25)
# ▶ (num 36)
# ▶ (num 49)
# ~> each {|x| put $x[..3] } [lorem ipsum]
# ▶ lor
# ▶ ips
# ```
#
# See also [`peach`]().
#
# Etymology: Various languages, as `for each`. Happens to have the same name as
# the iteration construct of
# [Factor](http://docs.factorcode.org/content/word-each,sequences.html).
fn each {|f inputs?| }

# Calls `$f` for each [value input](#value-inputs), possibly in parallel.
#
# Like `each`, an exception raised from [`break`]() will cause `peach` to
# terminate early. However due to the parallel nature of `peach`, the exact time
# of termination is non-deterministic, and termination is not guaranteed.
#
# An exception raised from [`continue`]() is swallowed and can be used to
# terminate a single iteration early.
#
# The `&num-workers` option restricts the number of functions that may run in
# parallel, and must be either an exact positive or `+inf`. A value of `+inf`
# (the default) means no restriction. Note that `peach &num-workers=1` is
# equivalent to `each`.
#
# Example (your output will differ):
#
# ```elvish-transcript
# //skip-test
# ~> range 1 10 | peach {|x| + $x 10 }
# ▶ (num 12)
# ▶ (num 13)
# ▶ (num 11)
# ▶ (num 16)
# ▶ (num 18)
# ▶ (num 14)
# ▶ (num 17)
# ▶ (num 15)
# ▶ (num 19)
# ~> range 1 101 |
#    peach {|x| if (== 50 $x) { break } else { put $x } } |
#    + (all) # 1+...+49 = 1225; 1+...+100 = 5050
# ▶ (num 1328)
# ```
#
# This command is intended for homogeneous processing of possibly unbound data. If
# you need to do a fixed number of heterogeneous things in parallel, use
# `run-parallel`.
#
# See also [`each`]() and [`run-parallel`]().
fn peach {|&num-workers=(num +inf) f inputs?| }

# Throws an exception; `$v` may be any type. If `$v` is already an exception,
# `fail` rethrows it.
#
# ```elvish-transcript
# ~> fail bad
# Exception: bad
#   [tty]:1:1-8: fail bad
# ~> put ?(fail bad)
# ▶ [^exception &reason=[^fail-error &content=bad &type=fail] &stack-trace=<...>]
# ~> fn f { fail bad }
# ~> fail ?(f)
# Exception: bad
#   [tty]:1:8-16: fn f { fail bad }
#   [tty]:1:8-8: fail ?(f)
# ```
fn fail {|v| }

# Raises the special "return" exception. When raised inside a named function
# (defined by the [`fn` keyword](language.html#fn)) it is captured by the
# function and causes the function to terminate. It is not captured by an
# ordinary anonymous function.
#
# Because `return` raises an exception it can be caught by a
# [`try`](language.html#try) block. If not caught, either implicitly by a
# named function or explicitly, it causes a failure like any other uncaught
# exception.
#
# See the discussion about [flow commands and
# exceptions](language.html#exception-and-flow-commands)
#
# **Note**: If you want to shadow the builtin `return` function with a local
# wrapper, do not define it with `fn` as `fn` swallows the special exception
# raised by return. Consider this example:
#
# ```elvish-transcript
# ~> use builtin
# ~> fn return { put return; builtin:return }
# ~> fn test-return { put before; return; put after }
# ~> test-return
# ▶ before
# ▶ return
# ▶ after
# ```
#
# Instead, shadow the function by directly assigning to `return~`:
#
# ```elvish-transcript
# ~> use builtin
# ~> var return~ = { put return; builtin:return }
# ~> fn test-return { put before; return; put after }
# ~> test-return
# ▶ before
# ▶ return
# ```
fn return { }

# Raises the special "break" exception. When raised inside a loop it is
# captured and causes the loop to terminate.
#
# Because `break` raises an exception it can be caught by a
# [`try`](language.html#try) block. If not caught, either implicitly by a loop
# or explicitly, it causes a failure like any other uncaught exception.
#
# See the discussion about [flow commands and exceptions](language.html#exception-and-flow-commands)
#
# **Note**: You can create a `break` function and it will shadow the builtin
# command. If you do so you should explicitly invoke the builtin. For example:
#
# ```elvish-transcript
# ~> use builtin
# ~> fn break { put 'break'; builtin:break; put 'should not appear' }
# ~> for x [a b c] { put $x; break; put 'unexpected' }
# ▶ a
# ▶ break
# ```
fn break { }

# Raises the special "continue" exception. When raised inside a loop it is
# captured and causes the loop to begin its next iteration.
#
# Because `continue` raises an exception it can be caught by a
# [`try`](language.html#try) block. If not caught, either implicitly by a loop
# or explicitly, it causes a failure like any other uncaught exception.
#
# See the discussion about [flow commands and exceptions](language.html#exception-and-flow-commands)
#
# **Note**: You can create a `continue` function and it will shadow the builtin
# command. If you do so you should explicitly invoke the builtin. For example:
#
# ```elvish-transcript
# ~> use builtin
# ~> fn continue { put 'continue'; builtin:continue; put 'should not appear' }
# ~> for x [a b c] { put $x; continue; put 'unexpected' }
# ▶ a
# ▶ continue
# ▶ b
# ▶ continue
# ▶ c
# ▶ continue
# ```
fn continue { }

# Schedules a function to be called when execution reaches the end of the
# current closure. The function is called with no arguments or options, and any
# exception it throws gets propagated.
#
# Examples:
#
# ```elvish-transcript
# ~> { defer { put foo }; put bar }
# ▶ bar
# ▶ foo
# ~> defer { put foo }
# Exception: defer must be called from within a closure
#   [tty]:1:1-17: defer { put foo }
# ```
fn defer {|fn| }
