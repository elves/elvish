# Starts custom listing mode.
#
# The `$items` argument can be as a list of maps, each map representing one item
# and having the following keys:
#
# - The value of the `to-show` key must be a string or a styled text. It is used
#   in the listing UI.
#
# - The value of the `to-filter` key must be a string. It is used when filtering
#   the item.
#
# - The value of the `to-accept` key must be a string. It is passed to the
#   accept callback (see below).
#
# Alternatively, the `$items` argument can be a function taking one argument. It
# will be called with the value of the filter (initially an empty string), and
# can output any number of maps containing the `to-show` and `to-accept` keys,
# with the same semantics as above. Any other key is ignored.
#
# The `&binding` option, if specified, should be a binding map to use in the
# custom listing mode. Bindings from [`$edit:listing:binding`]() are also used,
# after this map if it is specified.
#
# The `&caption` option changes the caption of the mode. If empty, the caption
# defaults to `' LISTING '`.
#
# The `&keep-bottom` option, if true, makes the last item to get selected
# initially or when the filter changes.
#
# The `&accept` option specifies a function to call when an item is accepted. It
# is passed the value of the `to-accept` key of the item.
#
# The `&auto-accept` option, if true, accepts an item automatically when there
# is only one item being shown.
fn listing:start-custom {|items &binding=$nil &caption='' &keep-bottom=$false &accept=$nil &auto-accept=$false| }
