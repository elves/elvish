# ```elvish
# edit:history:binding
# ```
#
# Binding table for the history mode.
var history:binding

# Starts the history mode.
fn history:start { }

# Walks to the previous entry in history mode.
fn history:up { }

# Walks to the next entry in history mode.
fn history:down { }

# Walks to the next entry in history mode, or quit the history mode if already
# at the newest entry.
fn history:down-or-quit { }

# Import command history entries that happened after the current session
# started.
fn history:fast-forward { }

# Replaces the content of the buffer with the current history mode entry, and
# closes history mode.
fn history:accept { }
