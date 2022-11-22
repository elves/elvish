# Moves the dot left one rune. Does nothing if the dot is at the beginning of
# the buffer.
fn move-dot-left { }

# Kills one rune left of the dot. Does nothing if the dot is at the beginning of
# the buffer.
fn kill-rune-left { }

# Moves the dot right one rune. Does nothing if the dot is at the end of the
# buffer.
fn move-dot-right { }

# Kills one rune right of the dot. Does nothing if the dot is at the end of the
# buffer.
fn kill-rune-left { }

# Moves the dot to the start of the current line.
fn move-dot-sol { }

# Deletes the text between the dot and the start of the current line.
fn kill-line-left { }

# Moves the dot to the end of the current line.
fn move-dot-eol { }

# Deletes the text between the dot and the end of the current line.
fn kill-line-right { }

# Moves the dot up one line, trying to preserve the visual horizontal position.
# Does nothing if dot is already on the first line of the buffer.
fn move-dot-up { }

# Moves the dot down one line, trying to preserve the visual horizontal
# position. Does nothing if dot is already on the last line of the buffer.
fn move-dot-down { }

# Swaps the runes to the left and right of the dot. If the dot is at the
# beginning of the buffer, swaps the first two runes, and if the dot is at the
# end, it swaps the last two.
fn transpose-rune { }

# Moves the dot to the beginning of the last word to the left of the dot.
fn move-dot-left-word { }

# Deletes the last word to the left of the dot.
fn kill-word-left { }

# Moves the dot to the beginning of the first word to the right of the dot.
fn move-dot-right-word { }

# Deletes the first word to the right of the dot.
fn kill-word-right { }

# Swaps the words to the left and right of the dot. If the dot is at the
# beginning of the buffer, swaps the first two words, and the dot is at the
# end, it swaps the last two.
fn transpose-word { }

# Moves the dot to the beginning of the last small word to the left of the dot.
fn move-dot-left-small-word { }

# Deletes the last small word to the left of the dot.
fn kill-small-word-left { }

# Moves the dot to the beginning of the first small word to the right of the dot.
fn move-dot-right-small-word { }

# Deletes the first small word to the right of the dot.
fn kill-small-word-right { }

# Swaps the small words to the left and right of the dot. If the dot is at the
# beginning of the buffer, it swaps the first two small words, and if the dot
# is at the end, it swaps the last two.
fn transpose-small-word { }

# Moves the dot to the beginning of the last alnum word to the left of the dot.
fn move-dot-left-alnum-word { }

# Deletes the last alnum word to the left of the dot.
fn kill-alnum-word-left { }

# Moves the dot to the beginning of the first alnum word to the right of the dot.
fn move-dot-right-alnum-word { }

# Deletes the first alnum word to the right of the dot.
fn kill-alnum-word-right { }

# Swaps the alnum words to the left and right of the dot. If the dot is at the
# beginning of the buffer, it swaps the first two alnum words, and if the dot
# is at the end, it swaps the last two.
fn transpose-alnum-word { }
