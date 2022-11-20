#elvdoc:fn cd
#
# ```elvish
# cd $dirname
# ```
#
# Changes directory.
#
# This affects the entire process, including parallel tasks that are started
# implicitly (such as prompt functions) or explicitly (such as one started by
# [`peach`](#peach)).
#
# Note that Elvish's `cd` does not support `cd -`.
#
# @cf pwd

#elvdoc:fn tilde-abbr
#
# ```elvish
# tilde-abbr $path
# ```
#
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
