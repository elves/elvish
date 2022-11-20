#elvdoc:fn move-dot-left
#
# ```elvish
# edit:move-dot-left
# ```
#
# Moves the dot left one rune. Does nothing if the dot is at the beginning of
# the buffer.

#elvdoc:fn kill-rune-left
#
# ```elvish
# edit:kill-rune-left
# ```
#
# Kills one rune left of the dot. Does nothing if the dot is at the beginning of
# the buffer.

#elvdoc:fn move-dot-right
#
# ```elvish
# edit:move-dot-right
# ```
#
# Moves the dot right one rune. Does nothing if the dot is at the end of the
# buffer.

#elvdoc:fn kill-rune-left
#
# ```elvish
# edit:kill-rune-left
# ```
#
# Kills one rune right of the dot. Does nothing if the dot is at the end of the
# buffer.

#elvdoc:fn move-dot-sol
#
# ```elvish
# edit:move-dot-sol
# ```
#
# Moves the dot to the start of the current line.

#elvdoc:fn kill-line-left
#
# ```elvish
# edit:kill-line-left
# ```
#
# Deletes the text between the dot and the start of the current line.

#elvdoc:fn move-dot-eol
#
# ```elvish
# edit:move-dot-eol
# ```
#
# Moves the dot to the end of the current line.

#elvdoc:fn kill-line-right
#
# ```elvish
# edit:kill-line-right
# ```
#
# Deletes the text between the dot and the end of the current line.

#elvdoc:fn move-dot-up
#
# ```elvish
# edit:move-dot-up
# ```
#
# Moves the dot up one line, trying to preserve the visual horizontal position.
# Does nothing if dot is already on the first line of the buffer.

#elvdoc:fn move-dot-down
#
# ```elvish
# edit:move-dot-down
# ```
#
# Moves the dot down one line, trying to preserve the visual horizontal
# position. Does nothing if dot is already on the last line of the buffer.

#elvdoc:fn transpose-rune
#
# ```elvish
# edit:transpose-rune
# ```
#
# Swaps the runes to the left and right of the dot. If the dot is at the
# beginning of the buffer, swaps the first two runes, and if the dot is at the
# end, it swaps the last two.

#elvdoc:fn move-dot-left-word
#
# ```elvish
# edit:move-dot-left-word
# ```
#
# Moves the dot to the beginning of the last word to the left of the dot.

#elvdoc:fn kill-word-left
#
# ```elvish
# edit:kill-word-left
# ```
#
# Deletes the last word to the left of the dot.

#elvdoc:fn move-dot-right-word
#
# ```elvish
# edit:move-dot-right-word
# ```
#
# Moves the dot to the beginning of the first word to the right of the dot.

#elvdoc:fn kill-word-right
#
# ```elvish
# edit:kill-word-right
# ```
#
# Deletes the first word to the right of the dot.

#elvdoc:fn transpose-word
#
# ```elvish
# edit:transpose-word
# ```
#
# Swaps the words to the left and right of the dot. If the dot is at the
# beginning of the buffer, swaps the first two words, and the dot is at the
# end, it swaps the last two.

#elvdoc:fn move-dot-left-small-word
#
# ```elvish
# edit:move-dot-left-small-word
# ```
#
# Moves the dot to the beginning of the last small word to the left of the dot.

#elvdoc:fn kill-small-word-left
#
# ```elvish
# edit:kill-small-word-left
# ```
#
# Deletes the last small word to the left of the dot.

#elvdoc:fn move-dot-right-small-word
#
# ```elvish
# edit:move-dot-right-small-word
# ```
#
# Moves the dot to the beginning of the first small word to the right of the dot.

#elvdoc:fn kill-small-word-right
#
# ```elvish
# edit:kill-small-word-right
# ```
#
# Deletes the first small word to the right of the dot.

#elvdoc:fn transpose-small-word
#
# ```elvish
# edit:transpose-small-word
# ```
#
# Swaps the small words to the left and right of the dot. If the dot is at the
# beginning of the buffer, it swaps the first two small words, and if the dot
# is at the end, it swaps the last two.

#elvdoc:fn move-dot-left-alnum-word
#
# ```elvish
# edit:move-dot-left-alnum-word
# ```
#
# Moves the dot to the beginning of the last alnum word to the left of the dot.

#elvdoc:fn kill-alnum-word-left
#
# ```elvish
# edit:kill-alnum-word-left
# ```
#
# Deletes the last alnum word to the left of the dot.

#elvdoc:fn move-dot-right-alnum-word
#
# ```elvish
# edit:move-dot-right-alnum-word
# ```
#
# Moves the dot to the beginning of the first alnum word to the right of the dot.

#elvdoc:fn kill-alnum-word-right
#
# ```elvish
# edit:kill-alnum-word-right
# ```
#
# Deletes the first alnum word to the right of the dot.

#elvdoc:fn transpose-alnum-word
#
# ```elvish
# edit:transpose-alnum-word
# ```
#
# Swaps the alnum words to the left and right of the dot. If the dot is at the
# beginning of the buffer, it swaps the first two alnum words, and if the dot
# is at the end, it swaps the last two.
