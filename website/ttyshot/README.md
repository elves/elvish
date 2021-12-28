This directory contains "ttyshots" -- they are like screenshots, but taken on
terminals. They are taken with Elvish's `edit:-dump-buf` function. To take one,
use the following procedure:

1.  Modify `edit:rprompt` to pretend that the username is `elf` and the hostname
    is `host`:

    ```elvish
    set edit:rprompt = (constantly (styled 'elf@host' inverse))
    ```

2.  Add a keybinding for taking ttyshots:

    ```elvish
    var header = '<!-- Follow website/ttyshot/README.md to regenerate -->'
    edit:global-binding[Alt-x] = { print $header(edit:-dump-buf) > ~/ttyshot.html }
    ```

3.  Make sure that the terminal width is 58, to be consistent with existing
    ttyshots.

4.  Put Elvish in the state you want, and press Alt-X. The ttyshot is saved at
    `~/ttyshot.html`.

Some of the ttyshots also show the output of commands. Since `edit:-dump-buf`
only captures the Elvish UI, you need to manually append the command output when
updating such ttyshots.
