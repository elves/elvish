There will be three classes of errors:

1. Static programming errors. The code doesn't pass the static checker and
   there is no way to handle it other than change the code.

2. Dynamic programming errors. Some pre-condition not enforced by the static
   checker is violated. Examples:

    1. `fn f a b {}; var args = [1]; f @$args`

    2. `/ 1 0`

3. Exceptional conditions that are anticipated but could not be reliably
   avoided. Examples:

    1. `echo >./out` but `./out` is not writable

    2. `ffmpeg -i a.mp4 a.ogg` but `a.mp4` is not valid MP4. Non-zero return
       values of external commands always fall into this category.
