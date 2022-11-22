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

# Moves the cursor left in listing mode.
fn listing:left { }

# Moves the cursor right in listing mode.
fn listing:right { }

# ```elvish
# edit:location:hidden
# ```
#
# A list of directories to hide in the location addon.
var location:hidden

# ```elvish
# edit:location:pinned
# ```
#
# A list of directories to always show at the top of the list of the location
# addon.
var location:pinned

# ```elvish
# edit:location:workspaces
# ```
#
# A map mapping types of workspaces to their patterns.
var location:workspaces
