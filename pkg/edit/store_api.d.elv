# Outputs the command history.
#
# By default, each entry is represented as a map, with an `id` key key for the
# sequence number of the command, and a `cmd` key for the text of the command.
# If `&cmd-only` is `$true`, only the text of each command is output.
#
# All entries are output by default. If `&dedup` is `$true`, only the most
# recent instance of each command (when comparing just the `cmd` key) is
# output.
#
# Commands are are output in oldest to newest order by default. If
# `&newest-first` is `$true` the output is in newest to oldest order instead.
#
# As an example, either of the following extracts the text of the most recent
# command:
#
# ```elvish
# edit:command-history | put [(all)][-1][cmd]
# edit:command-history &cmd-only &newest-first | take 1
# ```
fn command-history {|&cmd-only=$false &dedup=$false &newest-first| }

# Inserts the last word of the last command.
fn insert-last-word { }
