# Name of the currently selected file in navigation mode. $nil if not in
# navigation mode.
var navigation:selected-file

# Keybinding for the navigation mode.
#
# Keys bound to
# [edit:navigation:trigger-filter](#edit:navigation:trigger-filter) (Ctrl-F by
# default) and
# [edit:navigation:trigger-shown-hidden](#edit:navigation:trigger-shown-hidden)
# (Ctrl-H by default) will be shown in the navigation mode UI.
var navigation:binding

# Start the navigation mode.
fn navigation:start { }

# Inserts the selected filename.
fn navigation:insert-selected { }

# Inserts the selected filename and closes the navigation addon.
fn navigation:insert-selected-and-quit { }

# Toggles the filtering status of the navigation addon.
fn navigation:trigger-filter { }

# Toggles whether the navigation addon should be showing hidden files.
fn navigation:trigger-shown-hidden { }

# A list of 3 integers, used for specifying the width ratio of the 3 columns in
# navigation mode.
var navigation:width-ratio
