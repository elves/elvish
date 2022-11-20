#elvdoc:fn exec
#
# ```elvish
# exec $command? $args...
# ```
#
# Replace the Elvish process with an external `$command`, defaulting to
# `elvish`, passing the given arguments. This decrements `$E:SHLVL` before
# starting the new process.
#
# This command always raises an exception on Windows with the message "not
# supported on Windows".
