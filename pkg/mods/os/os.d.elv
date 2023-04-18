# Removes one or more path names.
#
# If passed zero path names it does nothing; otherwise, it iterates over the
# list of path names and attempts to remove each one. If a path name does not
# exist or cannot be removed (perhaps because it is a non-empty directory or
# permissions do not allow the operation) an exception is raised.
#
# Like the traditional Unix `rm` command an error processing a path name does
# not immediately terminate processing the list of path names. This command
# attempts to remove the remaining path names. This can result in a "multiple
# error" exception that documents each path that could not be removed.
#
# If the `&missing-ok` option is set to true a path name that does not
# exist is silently ignored.
#
# If the `&recursive` option is true a path name that refers to a directory is
# removed recursively. If this option is false then attempting to remove a
# directory that is not empty will fail.
#
# ```elvish-transcript
# ~> mkdir elv
# ~> mkdir elv/d
# ~> touch elv/a
# ~> touch elv/d/a
# ~> os:remove elv/x
# Exception: path does not exist: elv/x
# ~> os:remove &missing-okay elv/x
# ~> os:remove elv
# Exception: remove elv/d: directory not empty
# ~> os:remove &recursive elv
# ```
fn remove {|&missing-ok=$false &recursive=$false path...| }
