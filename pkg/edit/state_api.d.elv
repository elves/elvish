#elvdoc:fn insert-at-dot
#
# ```elvish
# edit:insert-at-dot $text
# ```
#
# Inserts the given text at the dot, moving the dot after the newly
# inserted text.

#elvdoc:fn replace-input
#
# ```elvish
# edit:replace-input $text
# ```
#
# Equivalent to assigning `$text` to `$edit:current-command`.

#elvdoc:var -dot
#
# Contains the current position of the cursor, as a byte position within
# `$edit:current-command`.

#elvdoc:var current-command
#
# Contains the content of the current input. Setting the variable will
# cause the cursor to move to the very end, as if `edit-dot = (count
# $edit:current-command)` has been invoked.
#
# This API is subject to change.
