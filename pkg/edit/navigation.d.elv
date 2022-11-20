#elvdoc:var navigation:selected-file
#
# Name of the currently selected file in navigation mode. $nil if not in
# navigation mode.

#elvdoc:var navigation:binding
#
# ```elvish
# edit:navigation:binding
# ```
#
# Keybinding for the navigation mode.

#elvdoc:fn navigation:start
#
# ```elvish
# edit:navigation:start
# ```
#
# Start the navigation mode.

#elvdoc:fn navigation:insert-selected
#
# ```elvish
# edit:navigation:insert-selected
# ```
#
# Inserts the selected filename.

#elvdoc:fn navigation:insert-selected-and-quit
#
# ```elvish
# edit:navigation:insert-selected-and-quit
# ```
#
# Inserts the selected filename and closes the navigation addon.

#elvdoc:fn navigation:trigger-filter
#
# ```elvish
# edit:navigation:trigger-filter
# ```
#
# Toggles the filtering status of the navigation addon.

#elvdoc:fn navigation:trigger-shown-hidden
#
# ```elvish
# edit:navigation:trigger-shown-hidden
# ```
#
# Toggles whether the navigation addon should be showing hidden files.

#elvdoc:var navigation:width-ratio
#
# A list of 3 integers, used for specifying the width ratio of the 3 columns in
# navigation mode.
