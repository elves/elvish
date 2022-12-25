# Converts a normal map into a binding map.
fn binding-table {|map| }

# Closes the current active mode.
fn close-mode { }

# Adds a notification saying "End of history".
fn end-of-history { }

# Triggers a redraw.
#
# The `&full` option controls whether to do a full redraw. By default, all
# redraws performed by the line editor are incremental redraws, updating only
# the part of the screen that has changed from the last redraw. A full redraw
# updates the entire command line.
fn redraw {|&full=$false| }

# Clears the screen.
#
# This command should be used in place of the external `clear` command to clear
# the screen.
fn clear { }

# Requests the next terminal input to be inserted uninterpreted.
fn insert-raw { }

# Parses a string into a key.
fn key {|string| }

# Prints a notification message. The argument may be a string or a [styled
# text](builtin.html#styled).
#
# If called while the editor is active, this will print the message above the
# editor, and redraw the editor.
#
# If called while the editor is inactive, the message will be queued, and shown
# once the editor becomes active.
fn notify {|message| }

# Causes the Elvish REPL to end the current read iteration and evaluate the
# code it just read. If called from a key binding, takes effect after the key
# binding returns.
fn return-line { }

# Causes the Elvish REPL to terminate. If called from a key binding, takes
# effect after the key binding returns.
fn return-eof { }

# If the current code is syntactically incomplete (like `echo [`), inserts a
# literal newline.
#
# Otherwise, applies any pending autofixes and accepts the current line.
fn smart-enter { }

# Breaks Elvish code into words.
fn wordify {|code| }
