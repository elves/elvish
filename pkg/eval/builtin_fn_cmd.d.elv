#//skip-test

# Construct a callable value for the external program `$program`. Example:
#
# ```elvish-transcript
# ~> var x = (external man)
# ~> $x ls # opens the manpage for ls
# ```
#
# See also [`has-external`]() and [`search-external`]().
fn external {|program| }

# Test whether `$command` names a valid external command. Examples (your output
# might differ):
#
# ```elvish-transcript
# ~> has-external cat
# ▶ $true
# ~> has-external lalala
# ▶ $false
# ```
#
# See also [`external`]() and [`search-external`]().
fn has-external {|command| }

# Output the full path of the external `$command`. Throws an exception when not
# found. Example (your output might vary):
#
# ```elvish-transcript
# ~> search-external cat
# ▶ /bin/cat
# ```
#
# See also [`external`]() and [`has-external`]().
fn search-external {|command| }

# Replace the Elvish process with an external `$command`, defaulting to
# `elvish`, passing the given arguments. This decrements `$E:SHLVL` before
# starting the new process.
#
# This command always raises an exception on Windows with the message "not
# supported on Windows".
fn exec {|command? @args| }

# Exit the Elvish process with `$status` (defaulting to 0).
fn exit {|status?| }
