<!-- toc -->

# Configuring the terminal to run Elvish

The recommended way to use Elvish as your default shell is to configure your
terminal to launch Elvish as the default command for new sessions.

## macOS terminals

<table>
  <tr>
    <th>Terminal</th>
    <th>Instructions</th>
  </tr>
  <tr>
    <td>Terminal.app</td>
    <td>
      Open <kbd>Terminal &gt; Preferences</kbd>.
      Ensure you are on the <kbd>Profiles</kbd> tab, which
      should be the default tab. In the right-hand panel, select the
      <kbd>Shell</kbd> tab. Tick
      <kbd>Run command</kbd>, put the path to Elvish in the
      textbox, and untick <kbd>Run inside shell</kbd>.
    </td>
  </tr>
  <tr>
    <td>iTerm2</td>
    <td>
      Open <kbd>iTerm &gt; Preferences</kbd>. Select the
      <kbd>Profiles</kbd> tab. In the right-hand panel under
      <kbd>Command</kbd>, change the dropdown from
      <kbd>Login Shell</kbd> to
      <kbd>Custom Shell</kbd>, and put the path to Elvish in the
      textbox.
    </td>
  </tr>
</table>

## Linux and BSD terminals

<table>
  <tr>
    <th>Terminal</th>
    <th>Instructions</th>
  </tr>
  <tr>
    <td>GNOME Terminal</td>
    <td>
      Open <kbd>Edit &gt; Preferences</kbd>. In the right-hand
      panel, select the <kbd>Command</kbd> tab, tick
      <kbd>Run a custom command instead of my shell</kbd>,
      and set <kbd>Custom command</kbd> to the path to Elvish.
    </td>
  </tr>
  <tr>
    <td>Konsole</td>
    <td>
      Open <kbd>Settings &gt; Edit Current Profile</kbd>.
      Set <kbd>Command</kbd> to the path to Elvish.
    </td>
  </tr>
  <tr>
    <td>XFCE Terminal</td>
    <td>
      Open <kbd>Edit &gt; Preferences</kbd>. Check
      <kbd>Run a custom command instead of my shell</kbd>,
      and set <kbd>Custom command</kbd> to the path to Elvish.
    </td>
  </tr>
  <tr class="table-section">
    <td colspan="2" class="notice">
      The following terminals only support a command-line flag for changing
      the shell
    </td>
  </tr>
  <tr>
    <td>LXTerminal</td>
    <td>Pass <code>--command $path_to_elvish</code>.</td>
  </tr>
  <tr>
    <td>rxvt</td>
    <td>Pass <code>-e $path_to_elvish</code>.</td>
  </tr>
  <tr>
    <td>xterm</td>
    <td>Pass <code>-e $path_to_elvish</code>.</td>
  </tr>
</table>

## tmux

Add the following to `~/.tmux.conf`:

```tmux
if-shell 'which elvish' 'set -g default-command elvish'
```

This only launches Elvish if it's available, so it's safe to have in a
`.tmux.conf` that you sync with machines where you haven't installed Elvish yet.

## Windows terminals

<table>
  <tr>
    <th>Terminal</th>
    <th>Instructions</th>
  </tr>
  <tr>
    <td>Windows Terminal</td>
    <td>
      Press <kbd>Ctrl</kbd>+<kbd>,</kbd> to
      open <i>Settings</i>. Go to <kbd>Add a new profile &gt; New
      empty profile</kbd>. Fill in the 'Name' and enter path to Elvish in
      the 'Command line' textbox. Go to <kbd>Startup</kbd>
      option and select Elvish as the 'Default profile'. Hit
      <kbd>Save</kbd>.
    </td>
  </tr>
  <tr>
    <td>ConEmu</td>
    <td>
      Press <kbd>Win</kbd>+<kbd>Alt</kbd>+
      <kbd>T</kbd> to open the <i>Startup Tasks</i> dialog.
      Click on <kbd>±</kbd> button to create a new task,
      give it Elvish alias, enter the path to Elvish in the 'Commands'
      textbox and tick the 'Default task for new console' checkbox.
      Click on <kbd>Save settings</kbd> to finish.
    </td>
  </tr>
</table>

## VS Code

Open the command palette and run "Open User Settings (JSON)". Add the following
to the JSON file:

```json
    "terminal.integrated.defaultProfile.linux": "elvish",
    "terminal.integrated.profiles.linux": {
        "elvish": {
            "path": "elvish"
        },
    }
```

Change `linux` to `osx` or `windows`, depending on your operating system. See
[VS Code's documentation](https://code.visualstudio.com/docs/terminal/profiles)
for more details.

# Changing your login shell

On Unix systems, you can also use Elvish as your login shell. Run the following
Elvish snippet:

```elvish
use runtime
if (not (has-value [(cat /etc/shells)] $runtime:elvish-path)) {
    echo $runtime:elvish-path | sudo tee -a /etc/shells
}
chsh -s $runtime:elvish-path
```

You can change your login shell back to the system default with `chsh -s ''`.

## Dealing with incompatible programs

Some programs assume that the user's login shell is a traditional POSIX-like
shell, so they won't work correctly if your login shell is Elvish. The following
programs are known to have issues:

-   GDB (see [#1795](https://b.elv.sh/1795))

Such programs usually rely on the `$SHELL` environment variable to query the
login shell, so you can override it to a POSIX shell, like the following:

```elvish
fn gdb {|@a|
  env SHELL=/bin/sh gdb $@a
}
```
