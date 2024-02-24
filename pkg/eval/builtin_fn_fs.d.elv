# Changes directory.
#
# This affects the entire process, including parallel tasks that are started
# implicitly (such as prompt functions) or explicitly (such as started by
# [`peach`]()).
#
# Note that Elvish's `cd` does not support `cd -`. You can also change to a
# specific directory by starting [Location
# Mode](../learn/tour.html#location-mode) and typing the desired path as a
# filter.
#
# See also [`$pwd`]() and [Location Mode](../learn/tour.html#location-mode).
fn cd {|dirname| }

# If `$path` represents a path under the home directory, replace the home
# directory with `~`. Examples:
#
# ```elvish-transcript
# ~> echo $E:HOME
# /Users/foo
# ~> tilde-abbr /Users/foo
# ▶ '~'
# ~> tilde-abbr /Users/foobar
# ▶ /Users/foobar
# ~> tilde-abbr /Users/foo/a/b
# ▶ '~/a/b'
# ```
fn tilde-abbr {|path| }
