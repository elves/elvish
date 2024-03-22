# Inserts the given text at the dot, moving the dot after the newly
# inserted text.
fn insert-at-dot {|text| }

# Equivalent to assigning `$text` to `$edit:current-command`.
fn replace-input {|text| }

#doc:show-unstable
# Contains the current position of the cursor, as a byte position within
# `$edit:current-command`.
var -dot

# Contains the content of the current input. Setting the variable will
# cause the cursor to move to the very end, as if `edit-dot = (count
# $edit:current-command)` has been invoked.
#
# This API is subject to change.
var current-command
