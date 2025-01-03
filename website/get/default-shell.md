<!-- toc -->

# Configuring the terminal to run Elvish

This is the recommended way to use Elvish as your default shell.

## macOS terminals

<table>
  <tr>
    <th>Terminal</th>
    <th>Instructions</th>
  </tr>
  <tr>
    <td>Terminal.app</td>
    <td>
      <ol>
        <li>Open <kbd>Terminal &gt; Preferences</kbd>.
        <li>Ensure you are on the <kbd>Profiles</kbd> tab (which should be the default tab).
        <li>In the right-hand panel, select the <kbd>Shell</kbd> tab.
        <li>Tick <kbd>Run command</kbd>, put the path to Elvish in the textbox next to it, and untick <kbd>Run inside shell</kbd>.
      </ol>
    </td>
  </tr>
  <tr>
    <td>iTerm2</td>
    <td>
      <ol>
        <li>Open <kbd>iTerm &gt; Preferences</kbd>.
        <li>Select the <kbd>Profiles</kbd> tab.
        <li>In the right-hand panel, change the dropdown under <strong>Command</strong> from <kbd>Login Shell</kbd> to either <kbd>Custom Shell</kbd> or <kbd>Command</kbd>, and put the path to Elvish in the textbox next to it.
      </ol>
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
      <ol>
        <li>Open <kbd>Edit &gt; Preferences</kbd>.
        <li>In the right-hand panel, select the <kbd>Command</kbd> tab.
        <li>Tick <kbd>Run a custom command instead of my shell</kbd>, and set <kbd>Custom command</kbd> to the path to Elvish.
      </ol>
    </td>
  </tr>
  <tr>
    <td>Konsole</td>
    <td>
      <ol>
        <li>Open <kbd>Settings &gt; Edit Current Profile</kbd>.
        <li>Set <kbd>Command</kbd> to the path to Elvish.
      </ol>
    </td>
  </tr>
  <tr>
    <td>XFCE Terminal</td>
    <td>
      <ol>
        <li>Open <kbd>Edit &gt; Preferences</kbd>.
        <li>Tick <kbd>Run a custom command instead of my shell</kbd>, and set <kbd>Custom command</kbd> to the path to Elvish.
      </ol>
    </td>
  </tr>
</table>

The following terminals only support a command-line flag to change the shell.
Depending on your DE, you can either create a wrapper script or
[modify the desktop file](https://wiki.archlinux.org/title/desktop_entries#Modify_desktop_files):

<table>
  <tr>
    <th>Terminal</th>
    <th>Instructions</th>
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
      <ol>
        <li>Press <kbd>Ctrl</kbd>+<kbd>,</kbd> to open <strong>Settings</strong>.
        <li>Select <kbd>Add a new profile</kbd> from the left sidebar, and click <kbd>New empty profile</kbd>.
        <li>Set <kbd>Name</kbd> to “Elvish” and <kbd>Command line</kbd> to the path to Elvish.
        <li>Select <kbd>Startup</kbd> from the left sidebar, and set <kbd>Default profile</kbd> to Elvish.
        <li>Hit <kbd>Save</kbd>.
      </ol>
    </td>
  </tr>
  <tr>
    <td>ConEmu</td>
    <td>
      <ol>
        <li>Press <kbd>Win</kbd>+<kbd>Alt</kbd>+<kbd>T</kbd> to open <strong>Startup Tasks</strong>.
        <li>Click <kbd>±</kbd> below the list of existing tasks.
        <li>Set the name to “Elvish”, enter the path to Elvish in the textbox below <strong>Commands</strong>, and tick <kbd>Default task for new console</kbd>.
        <li>Click <kbd>Save settings</kbd>.
      </ol>
    </td>
  </tr>
</table>

## VS Code

Open the command palette and run "Open User Settings (JSON)". Add the following:

```json
    "terminal.integrated.defaultProfile.linux": "elvish",
    "terminal.integrated.profiles.linux": {
        "elvish": {
            "path": "elvish"
        },
    }
```

Change `linux` to `osx` or `windows` depending on your operating system. See
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

Some programs invoke the user's login shell assuming that it is a traditional
POSIX-like shell, so they won't work correctly if your login shell is Elvish.
This section lists programs known to have issues and possible workarounds.

If you can't work around the issue, it may be easier to switch the login shell
back to the system default and configure your terminal to launch Elvish instead.

### Vim / Neovim

Add the following to `.vimrc` or `.nvimrc`:

```
set shell=/bin/sh
```

### GDB

Create an alias that sets the `SHELL` environment variable to `/bin/sh`:

```elvish
fn gdb {|@a|
  env SHELL=/bin/sh gdb $@a
}
```

(Reported in [#1795](https://b.elv.sh/1795).)

### vscode-neovim

No workaround has been tested yet. Launching VS Code with the `SHELL`
environment variable set to `/bin/sh` may work.

If you use vscode-neovim and have found and tested a workaround, please send a
pull request updating this document.

(Reported in [#1804](https://b.elv.sh/1804).)
