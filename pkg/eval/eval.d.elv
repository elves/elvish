#//skip-test
# A list of functions to run after changing directory. These functions are always
# called with directory to change it, which might be a relative path. The
# following example also shows `$before-chdir`:
#
# ```elvish-transcript
# ~> set before-chdir = [{|dir| echo "Going to change to "$dir", pwd is "$pwd }]
# ~> set after-chdir = [{|dir| echo "Changed to "$dir", pwd is "$pwd }]
# ~> cd /usr
# Going to change to /usr, pwd is /Users/xiaq
# Changed to /usr, pwd is /usr
# /usr> cd local
# Going to change to local, pwd is /usr
# Changed to local, pwd is /usr/local
# /usr/local>
# ```
#
# **Note**: The use of `echo` above is for illustrative purposes. When Elvish
# is used interactively, the working directory may be changed in location mode
# or navigation mode, and outputs from `echo` can garble the terminal. If you
# are writing a plugin that works with the interactive mode, it's better to use
# [`edit:notify`](edit.html#edit:notify).
#
# See also [`$before-chdir`]().
var after-chdir

# A list of functions to run before changing directory. These functions are always
# called with the new working directory.
#
# See also [`$after-chdir`]().
var before-chdir

# A list of functions to run before Elvish exits.
var before-exit

# Number of background jobs.
var num-bg-jobs

# Whether to notify success of background jobs, defaulting to `$true`.
#
# Failures of background jobs are always notified.
var notify-bg-job-success

#//skip-test
#// The test framework hardcodes value out indicators.
# A string put before value outputs (such as those of `put`). Defaults to
# `'▶ '`. Example:
#
# ```elvish-transcript
# ~> put lorem ipsum
# ▶ lorem
# ▶ ipsum
# ~> set value-out-indicator = 'val> '
# ~> put lorem ipsum
# val> lorem
# val> ipsum
# ```
#
# Note that you almost always want some trailing whitespace for readability.
var value-out-indicator
