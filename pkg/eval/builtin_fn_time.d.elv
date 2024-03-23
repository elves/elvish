#//skip-test

# Pauses for at least the specified duration. The actual pause duration depends
# on the system.
#
# This only affects the current Elvish context. It does not affect any other
# contexts that might be executing in parallel as a consequence of a command
# such as [`peach`]().
#
# A duration can be a simple [number](language.html#number) (with optional
# fractional value) without an explicit unit suffix, with an implicit unit of
# seconds.
#
# A duration can also be a string written as a sequence of decimal numbers,
# each with optional fraction, plus a unit suffix. For example, "300ms",
# "1.5h" or "1h45m7s". Valid time units are "ns", "us" (or "µs"), "ms", "s",
# "m", "h".
#
# Passing a negative duration causes an exception; this is different from the
# typical BSD or GNU `sleep` command that silently exits with a success status
# without pausing when given a negative duration.
#
# See the [Go documentation](https://golang.org/pkg/time/#ParseDuration) for
# more information about how durations are parsed.
#
# Examples:
#
# ```elvish-transcript
# ~> sleep 0.1    # sleeps 0.1 seconds
# ~> sleep 100ms  # sleeps 0.1 seconds
# ~> sleep 1.5m   # sleeps 1.5 minutes
# ~> sleep 1m30s  # sleeps 1.5 minutes
# ~> sleep -1
# Exception: sleep duration must be >= zero
# [tty 8], line 1: sleep -1
# ```
fn sleep {|duration| }

# Runs the callable, and call `$on-end` with the duration it took, as a
# number in seconds. If `$on-end` is `$nil` (the default), prints the
# duration in human-readable form.
#
# If `$callable` throws an exception, the exception is propagated after the
# on-end or default printing is done.
#
# If `$on-end` throws an exception, it is propagated, unless `$callable` has
# already thrown an exception.
#
# Example:
#
# ```elvish-transcript
# ~> time { sleep 1 }
# 1.006060647s
# ~> time { sleep 0.01 }
# 1.288977ms
# ~> var t = ''
# ~> time &on-end={|x| set t = $x } { sleep 1 }
# ~> put $t
# ▶ (num 1.000925004)
# ~> time &on-end={|x| set t = $x } { sleep 0.01 }
# ~> put $t
# ▶ (num 0.011030208)
# ```
#
# See also [`benchmark`]().
fn time {|&on-end=$nil callable| }

# Runs `$callable` repeatedly, and reports statistics about how long each run
# takes.
#
# If the `&on-end` callback is not given, `benchmark` prints the average,
# standard deviation, minimum and maximum of the time it took to run
# `$callback`, and the number of runs. If the `&on-end` callback is given,
# `benchmark` instead calls it with a map containing these metrics, keyed by
# `avg`, `stddev`, `min`, `max` and `runs`. Each duration value (i.e. all
# except `runs`) is given as the number of seconds.
#
# The number of runs is controlled by `&min-runs` and `&min-time`. The
# `$callable` is run at least `&min-runs` times, **and** when the total
# duration is at least `&min-time`.
#
# The `&min-runs` option must be a non-negative integer within the range of the
# machine word.
#
# The `&min-time` option must be a string representing a non-negative duration,
# specified as a sequence of decimal numbers with a unit suffix (the numbers
# may have fractional parts), such as "300ms", "1.5h" and "1h45m7s". Valid time
# units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
#
# If `&on-run-end` is given, it is called after each call to `$callable`, with
# the time that call took, given as the number of seconds.
#
# If `$callable` throws an exception, `benchmark` terminates and propagates the
# exception after the `&on-end` callback (or the default printing behavior)
# finishes. The duration of the call that throws an exception is not passed to
# `&on-run-end`, nor is it included when calculating the statistics for
# `&on-end`. If the first call to `$callable` throws an exception and `&on-end`
# is `$nil`, nothing is printed and any `&on-end` callback is not called.
#
# If `&on-run-end` is given and throws an exception, `benchmark` terminates and
# propagates the exception after the `&on-end` callback (or the default
# printing behavior) finishes, unless `$callable` has already thrown an
# exception
#
# If `&on-end` throws an exception, the exception is propagated, unless
# `$callable` or `&on-run-end` has already thrown an exception.
#
# Example:
#
# ```elvish-transcript
# ~> benchmark { }
# 98ns ± 382ns (min 0s, max 210.417µs, 10119226 runs)
# ~> benchmark &on-end={|m| put $m[avg]} { }
# ▶ (num 9.8e-08)
# ~> benchmark &on-run-end={|d| echo $d} { sleep 0.3 }
# 0.301123625
# 0.30123775
# 0.30119075
# 0.300629166
# 0.301260333
# 301.088324ms ± 234.298µs (min 300.629166ms, max 301.260333ms, 5 runs)
# ```
#
# See also [`time`]().
fn benchmark {|&min-runs=5 &min-time=1s &on-end=$nil &on-run-end=$nil callable| }
