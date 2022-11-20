#elvdoc:fn external
#
# ```elvish
# external $program
# ```
#
# Construct a callable value for the external program `$program`. Example:
#
# ```elvish-transcript
# ~> var x = (external man)
# ~> $x ls # opens the manpage for ls
# ```
#
# @cf has-external search-external

#elvdoc:fn has-external
#
# ```elvish
# has-external $command
# ```
#
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
# @cf external search-external

#elvdoc:fn search-external
#
# ```elvish
# search-external $command
# ```
#
# Output the full path of the external `$command`. Throws an exception when not
# found. Example (your output might vary):
#
# ```elvish-transcript
# ~> search-external cat
# ▶ /bin/cat
# ```
#
# @cf external has-external

#elvdoc:fn exit
#
# ```elvish
# exit $status?
# ```
#
# Exit the Elvish process with `$status` (defaulting to 0).
