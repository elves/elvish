# Maximum height the editor is allowed to use, defaults to `+Inf`.
#
# By default, the height of the editor is only restricted by the terminal
# height. Some modes like location mode can use a lot of lines; as a result,
# it can often occupy the entire terminal, and push up your scrollback buffer.
# Change this variable to a finite number to restrict the height of the editor.
var max-height

# A list of functions to call before each readline cycle. Each function is
# called without any arguments.
var before-readline

# A list of functions to call after each readline cycle. Each function is
# called with a single string argument containing the code that has been read.
var after-readline

# List of filters to run before adding a command to history.
#
# A filter is a function that takes a command as argument and outputs
# a boolean value. If any of the filters outputs `$false`, the
# command is not saved to history, and the rest of the filters are
# not run. The default value of this list contains a filter which
# ignores command starts with space.
var add-cmd-filters

# Global keybindings, consulted for keys not handled by mode-specific bindings.
#
# See [Keybindings](#keybindings).
var global-binding
