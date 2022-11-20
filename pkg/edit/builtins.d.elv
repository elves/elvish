#elvdoc:fn binding-table
#
# ```elvish
# edit:binding-table $map
# ```
#
# Converts a normal map into a binding map.

#elvdoc:fn close-mode
#
# ```elvish
# edit:close-mode
# ```
#
# Closes the current active mode.

#elvdoc:fn end-of-history
#
# ```elvish
# edit:end-of-history
# ```
#
# Adds a notification saying "End of history".

#elvdoc:fn redraw
#
# ```elvish
# edit:redraw &full=$false
# ```
#
# Triggers a redraw.
#
# The `&full` option controls whether to do a full redraw. By default, all
# redraws performed by the line editor are incremental redraws, updating only
# the part of the screen that has changed from the last redraw. A full redraw
# updates the entire command line.

#elvdoc:fn clear
#
# ```elvish
# edit:clear
# ```
#
# Clears the screen.
#
# This command should be used in place of the external `clear` command to clear
# the screen.

#elvdoc:fn insert-raw
#
# ```elvish
# edit:insert-raw
# ```
#
# Requests the next terminal input to be inserted uninterpreted.

#elvdoc:fn key
#
# ```elvish
# edit:key $string
# ```
#
# Parses a string into a key.

#elvdoc:fn notify
#
# ```elvish
# edit:notify $message
# ```
#
# Prints a notification message. The argument may be a string or a [styled
# text](builtin.html#styled).
#
# If called while the editor is active, this will print the message above the
# editor, and redraw the editor.
#
# If called while the editor is inactive, the message will be queued, and shown
# once the editor becomes active.

#elvdoc:fn return-line
#
# ```elvish
# edit:return-line
# ```
#
# Causes the Elvish REPL to end the current read iteration and evaluate the
# code it just read. If called from a key binding, takes effect after the key
# binding returns.

#elvdoc:fn return-eof
#
# ```elvish
# edit:return-eof
# ```
#
# Causes the Elvish REPL to terminate. If called from a key binding, takes
# effect after the key binding returns.

#elvdoc:fn smart-enter
#
# ```elvish
# edit:smart-enter
# ```
#
# Inserts a literal newline if the current code is not syntactically complete
# Elvish code. Accepts the current line otherwise.

#elvdoc:fn wordify
#
# ```elvish
# edit:wordify $code
# ```
#
# Breaks Elvish code into words.
