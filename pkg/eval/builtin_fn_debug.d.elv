#//skip-test

# Output a map describing the current source, which is the source file or
# interactive command that contains the call to `src`. The value contains the
# following fields:
#
# -   `name`, a unique name of the current source. If the source originates from a
#     file, it is the full path of the file.
#
# -   `code`, the full body of the current source.
#
# -   `is-file`, whether the source originates from a file.
#
# Examples:
#
# ```elvish-transcript
# ~> src
# ▶ [&code=src &is-file=$false &name='[tty 1]']
# ~> elvish show-src.elv
# ▶ [&code="src\n" &is-file=$true &name=/home/elf/show-src.elv]
# ~> echo src > .config/elvish/lib/show-src.elv
# ~> use show-src
# ▶ [&code="src\n" &is-file=$true &name=/home/elf/.config/elvish/lib/show-src.elv]
# ```
fn src { }

#doc:show-unstable
# Force the Go garbage collector to run.
#
# This is only useful for debug purposes.
fn -gc { }

#doc:show-unstable
# Print a stack trace.
#
# This is only useful for debug purposes.
fn -stack { }

#doc:show-unstable
# Direct internal debug logs to the named file.
#
# This is only useful for debug purposes.
fn -log {|filename| }
