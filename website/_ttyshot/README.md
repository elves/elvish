This directory contains "ttyshots" -- they are like screenshots, but taken on
terminals. They are taken with Elvish's undocumented `edit:-dump-buf` function.
To take one, use the following procedure:

1.  Modify `edit:rprompt` to pretend that the username is `elf` and the hostname
    is `host`:

    ```elvish
    edit:rprompt = (constantly (styled 'elf@host' inverse))
    ```

2.  Add a keybinding for taking ttyshots:

    ```elvish
    edit:insert:binding[Alt-x] = { edit:-dump-buf > ~/ttyshot.html }
    ```

3.  Make sure that the terminal width is 58, to be consistent with existing
    ttyshots.

4.  Put Elvish in the state you want, and press Alt-X. The ttyshot is saved at
    `~/ttyshot.html`.
