#elvdoc:fn next-cmd-seq
#
# ```elvish
# store:next-cmd-seq
# ```
#
# Outputs the sequence number that will be used for the next entry of the
# command history.

#elvdoc:fn add-cmd
#
# ```elvish
# store:add-cmd $text
# ```
#
# Adds an entry to the command history with the given content. Outputs its
# sequence number.

#elvdoc:fn del-cmd
#
# ```elvish
# store:del-cmd $seq
# ```
#
# Deletes the command history entry with the given sequence number.
#
# **NOTE**: This command only deletes the entry from the persistent store. When
# deleting an entry that was added in the current session, the deletion will
# not take effect for the current session, since the entry still exists in the
# in-memory per-session history.

#elvdoc:fn cmd
#
# ```elvish
# store:cmd $seq
# ```
#
# Outputs the content of the command history entry with the given sequence
# number.

#elvdoc:fn cmds
#
# ```elvish
# store:cmds $from $upto
# ```
#
# Outputs all command history entries with sequence numbers between `$from`
# (inclusive) and `$upto` (exclusive). Use -1 for `$upto` to not set an upper
# bound.
#
# Each entry is represented by a pseudo-map with fields `text` and `seq`.

#elvdoc:fn add-dir
#
# ```elvish
# store:add-dir $path
# ```
#
# Adds a path to the directory history. This will also cause the scores of all
# other directories to decrease.

#elvdoc:fn del-dir
#
# ```elvish
# store:del-dir $path
# ```
#
# Deletes a path from the directory history. This has no impact on the scores
# of other directories.

#elvdoc:fn dirs
#
# ```elvish
# store:dirs
# ```
#
# Outputs all directory history entries, in decreasing order of score.
#
# Each entry is represented by a pseudo-map with fields `path` and `score`.
