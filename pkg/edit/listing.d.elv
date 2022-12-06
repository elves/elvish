# Common binding table for [listing modes](#listing-modes).
var listing:binding { }

# Accepts the current selected listing item.
fn listing:accept { }

# Moves the cursor up in listing mode.
fn listing:up { }

# Moves the cursor down in listing mode.
fn listing:down { }

# Moves the cursor up in listing mode, or to the last item if the first item is
# currently selected.
fn listing:up-cycle { }

# Moves the cursor down in listing mode, or to the first item if the last item is
# currently selected.
fn listing:down-cycle { }

# Moves the cursor up one page.
fn listing:page-up { }

# Moves the cursor down one page.
fn listing:page-down { }

# Starts the history listing mode.
fn histlist:start { }

# Toggles deduplication in history listing mode.
#
# When deduplication is on (the default), only the last occurrence of the same
# command is shown.
fn histlist:toggle-dedup { }

# Keybinding for the history listing mode.
#
# Keys bound to [edit:histlist:toggle-dedup](#edit:histlist:toggle-dedup)
# (Ctrl-D by default) will be shown in the history listing UI.
var histlist:binding

# Starts the last command mode.
fn lastcmd:start { }

# Keybinding for the last command mode.
var lastcmd:binding

# Starts the location mode.
fn location:start

# Keybinding for the location mode.
var location:binding

# A list of directories to hide in the location addon.
var location:hidden

# A list of directories to always show at the top of the list of the location
# addon.
var location:pinned

# A map mapping types of workspaces to their patterns.
var location:workspaces
