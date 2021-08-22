<!-- toc -->

# Installing an official binary

The recommended way to install Elvish is by downloading an official binary.

First, choose the version to install. At any given time, two versions of Elvish
are supported:

-   The HEAD version tracks the latest development, and is updated shortly after
    every commit.

    Use HEAD if you want to use the latest features, and can live with
    occasional bugs and breaking changes.

-   The release version is updated with new features every 6 months, and gets
    occasional patch releases that fix severe issues.

    Use the release version if you want a stable foundation. You still need to
    update when a new release comes out, since only the latest release is
    supported.

Now find your platform in the table, and download the corresponding binary
archive:

<table>
  <tr>
    <th>Version</th>
    <th>amd64</th>
    <th>386</th>
    <th>arm64</th>
  </tr>
  <tr>
    <td>HEAD (<a href="https://github.com/elves/elvish/blob/master/0.17.0-release-notes.md">Draft Release Note</a>)</td>
    <td>
      @dl Linux linux-amd64/elvish-HEAD.tar.gz
      @dl macOS darwin-amd64/elvish-HEAD.tar.gz
      @dl FreeBSD freebsd-amd64/elvish-HEAD.tar.gz
      @dl NetBSD netbsd-amd64/elvish-HEAD.tar.gz
      @dl OpenBSD openbsd-amd64/elvish-HEAD.tar.gz
      @dl Windows windows-amd64/elvish-HEAD.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-HEAD.tar.gz
      @dl Windows windows-386/elvish-HEAD.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-HEAD.tar.gz
      @dl macOS darwin-arm64/elvish-HEAD.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.16.1 (<a href="../blog/0.16.0-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.16.1.tar.gz
      @dl macOS darwin-amd64/elvish-v0.16.1.tar.gz
      @dl FreeBSD freebsd-amd64/elvish-v0.16.1.tar.gz
      @dl NetBSD netbsd-amd64/elvish-v0.16.1.tar.gz
      @dl OpenBSD openbsd-amd64/elvish-v0.16.1.tar.gz
      @dl Windows windows-amd64/elvish-v0.16.1.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.16.1.tar.gz
      @dl Windows windows-386/elvish-v0.16.1.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.16.1.tar.gz
      @dl macOS darwin-arm64/elvish-v0.16.1.tar.gz
    </td>
  </tr>
</table>

(If your platform is not listed, you may still be able to build Elvish from
[source](https://github.com/elves/elvish). For users in China,
[TUNA's mirror](https://mirrors.tuna.tsinghua.edu.cn/elvish) may be faster.)

After downloading the binary archive, following these steps to install it:

```elvish
cd ~/Downloads # or wherever the binary archive was downloaded to
tar xvf elvish-HEAD.tar.gz # or elvish-v0.15.0.tar.gz for release version
chmod +x elvish-HEAD # or elvish-v0.15.0 for release version
sudo cp elvish-HEAD /usr/local/bin/elvish # or anywhere else on PATH
```

On Windows, simply unzip the downloaded archive and move it to the desktop. If
additionally you'd like to invoke `elvish` from `cmd`, move it to somewhere in
the `PATH` instead and create a desktop shortcut.

# Using Elvish as your default shell

On non-Windows systems, the best way to use Elvish as your default shell is to
configure your terminal to launch Elvish:

<table>
  <tr>
    <th>Terminal</th>
    <th>Instructions</th>
  </tr>
  <tr class="table-section">
    <td colspan="2" class="notice">
      Terminals for macOS
    </td>
  </tr>
  <tr>
    <td>Terminal.app</td>
    <td>
      Open <span class="key">Terminal &gt; Preferences</span>.
      Ensure you are on the <span class="key">Profiles</span> tab, which
      should be the default tab. In the right-hand panel, select the
      <span class="key">Shell</span> tab. Tick
      <span class="key">Run command</span>, put the path to Elvish in the
      textbox, and untick <span class="key">Run inside shell</span>.
    </td>
  </tr>
  <tr>
    <td>iTerm2</td>
    <td>
      Open <span class="key">iTerm &gt; Preferences</span>. Select the
      <span class="key">Profiles</span> tab. In the right-hand panel under
      <span class="key">Command</span>, change the dropdown from
      <span class="key">Login Shell</span> to
      <span class="key">Custom Shell</span>, and put the path to Elvish in the
      textbox.
    </td>
  </tr>
  <tr class="table-section">
    <td colspan="2" class="notice">
      Terminals for Linux and BSDs
    </td>
  </tr>
  <tr>
    <td>GNOME Terminal</td>
    <td>
      Open <span class="key">Edit &gt; Preferences</span>. In the right-hand
      panel, select the <span class="key">Command</span> tab, tick
      <span class="key">Run a custom command instead of my shell</span>,
      and set <span class="key">Custom command</span> to the path to Elvish.
    </td>
  </tr>
  <tr>
    <td>Konsole</td>
    <td>
      Open <span class="key">Settings &gt; Edit Current Profile</span>.
      Set <span class="key">Command</span> to the path to Elvish.
    </td>
  </tr>
  <tr>
    <td>XFCE Terminal</td>
    <td>
      Open <span class="key">Edit &gt; Preferences</span>. Check
      <span class="key">Run a custom command instead of my shell</span>,
      and set <span class="key">Custom command</span> to the path to Elvish.
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
  <tr class="table-section">
    <td colspan="2" class="notice">
      Terminal multiplexers
    </td>
  </tr>
  <tr>
    <td>tmux</td>
    <td>
      Add <code>set -g default-command $path_to_elvish</code> to
      <code>~/.tmux.conf</code>.
    </td>
  </tr>
</table>

It is **not** recommended to change your login shell to Elvish. Some programs
assume that user's login shell is a traditional POSIX-like shell, and may have
issues when you change your login shell to Elvish.

# Installing from a package manager

Elvish is available from many package managers. Installing Elvish with the
package manager makes it easy to upgrade Elvish alongside the rest of your
system.

Beware that these packages are not maintained by Elvish developers and are
sometimes out of date. For a comprehensive list of packages and their freshness,
see [this Repology page](https://repology.org/project/elvish/versions).

## Arch Linux

Elvish is available in the official repository. This will install the latest
release:

```elvish
pacman -S elvish
```

To install the HEAD version, install
[`elvish-git`](https://aur.archlinux.org/packages/elvish-git/) from AUR with
your favorite AUR helper:

```elvish
yay -S elvish-git
```

## Fedora

RPM packages are available from
[the FZUG Repo](https://github.com/FZUG/repo/wiki/Add-FZUG-Repository):

```elvish
# Add FZUG repo
dnf config-manager --add-repo=http://repo.fdzh.org/FZUG/FZUG.repo
# Install Elvish
dnf install elvish
```

## Debian / Ubuntu

Elvish is packaged by [Debian](https://packages.debian.org/elvish) since buster
and by [Ubuntu](http://packages.ubuntu.com/elvish) since 17.10:

```elvish
apt install elvish
```

However, only testing versions of Debian and Ubuntu tend to have the latest
Elvish release. If you are running a stable release of Debian or Ubuntu, it is
recommended to use official [prebuilt binaries](#prebuilt-binaries) instead.

## macOS

Elvish is packaged by both [Homebrew](https://brew.sh) and
[MacPorts](https://www.macports.org).

To install from Homebrew:

```elvish
# Install latest release
brew install elvish
# Or install HEAD:
brew install --HEAD elvish
```

To install from MacPorts:

```elvish
sudo port selfupdate
sudo port install elvish
```

## FreeBSD

Elvish is available in the FreeBSD ports tree and as a prebuilt package. Both
methods will install the latest release.

To install with `pkg`:

```elvish
pkg install elvish
```

To build from the ports tree:

```elvish
cd /usr/ports/shells/elvish
make install
```

## OpenBSD

Elvish is available in the official OpenBSD package repository. This will
install the latest release:

```elvish
doas pkg_add elvish
```

## NixOS (nix)

Elvish is packaged in
[nixpkgs](https://github.com/NixOS/nixpkgs/blob/master/pkgs/shells/elvish/default.nix):

```elvish
# Install latest release
nix-env -i elvish
```

# Old versions

The following old versions are no longer supported. They are only listed here
for historical interest.

<table>
  <tr>
    <th>Version</th>
    <th>amd64</th>
    <th>386</th>
    <th>arm64</th>
  </tr>
  <tr>
    <td>
      0.16.0 (<a href="../blog/0.16.0-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.16.0.tar.gz
      @dl macOS darwin-amd64/elvish-v0.16.0.tar.gz
      @dl FreeBSD freebsd-amd64/elvish-v0.16.0.tar.gz
      @dl NetBSD netbsd-amd64/elvish-v0.16.0.tar.gz
      @dl OpenBSD openbsd-amd64/elvish-v0.16.0.tar.gz
      @dl Windows windows-amd64/elvish-v0.16.0.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.16.0.tar.gz
      @dl Windows windows-386/elvish-v0.16.0.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.16.0.tar.gz
      @dl macOS darwin-arm64/elvish-v0.16.0.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.15.0 (<a href="../blog/0.15.0-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.15.0.tar.gz
      @dl macOS darwin-amd64/elvish-v0.15.0.tar.gz
      @dl FreeBSD freebsd-amd64/elvish-v0.15.0.tar.gz
      @dl NetBSD netbsd-amd64/elvish-v0.15.0.tar.gz
      @dl OpenBSD openbsd-amd64/elvish-v0.15.0.tar.gz
      @dl Windows windows-amd64/elvish-v0.15.0.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.15.0.tar.gz
      @dl Windows windows-386/elvish-v0.15.0.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.15.0.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.14.1 (<a href="../blog/0.14.1-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.14.1.tar.gz
      @dl macOS darwin-amd64/elvish-v0.14.1.tar.gz
      @dl FreeBSD freebsd-amd64/elvish-v0.14.1.tar.gz
      @dl NetBSD netbsd-amd64/elvish-v0.14.1.tar.gz
      @dl OpenBSD openbsd-amd64/elvish-v0.14.1.tar.gz
      @dl Windows windows-amd64/elvish-v0.14.1.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.14.1.tar.gz
      @dl Windows windows-386/elvish-v0.14.1.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.14.1.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.14.0 (<a href="/blog/0.14.0-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.14.0.tar.gz
      @dl macOS darwin-amd64/elvish-v0.14.0.tar.gz
      @dl FreeBSD freebsd-amd64/elvish-v0.14.0.tar.gz
      @dl NetBSD netbsd-amd64/elvish-v0.14.0.tar.gz
      @dl OpenBSD openbsd-amd64/elvish-v0.14.0.tar.gz
      @dl Windows windows-amd64/elvish-v0.14.0.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.14.0.tar.gz
      @dl Windows windows-386/elvish-v0.14.0.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.14.0.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.13.1 (<a href="/blog/0.13.1-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.13.1.tar.gz
      @dl macOS darwin-amd64/elvish-v0.13.1.tar.gz
      @dl FreeBSD freebsd-amd64/elvish-v0.13.1.tar.gz
      @dl NetBSD netbsd-amd64/elvish-v0.13.1.tar.gz
      @dl OpenBSD openbsd-amd64/elvish-v0.13.1.tar.gz
      @dl Windows windows-amd64/elvish-v0.13.1.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.13.1.tar.gz
      @dl Windows windows-386/elvish-v0.13.1.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.13.1.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.13 (<a href="/blog/0.13-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.13.tar.gz
      @dl macOS darwin-amd64/elvish-v0.13.tar.gz
      @dl Windows windows-amd64/elvish-v0.13.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.13.tar.gz
      @dl Windows windows-386/elvish-v0.13.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.13.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.12 (<a href="/blog/0.12-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.12.tar.gz
      @dl macOS darwin-amd64/elvish-v0.12.tar.gz
      @dl Windows windows-amd64/elvish-v0.12.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.12.tar.gz
      @dl Windows windows-386/elvish-v0.12.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.12.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.11 (<a href="/blog/0.11-release-notes.html">Release Note</a>)
    </td>
    <td>
      @dl Linux linux-amd64/elvish-v0.11.tar.gz
      @dl macOS darwin-amd64/elvish-v0.11.tar.gz
      @dl Windows windows-amd64/elvish-v0.11.zip
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.11.tar.gz
      @dl Windows windows-386/elvish-v0.11.zip
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.11.tar.gz
    </td>
  </tr>
  <tr>
    <td colspan="4" class="notice">
      Versions before 0.11 do not build on Windows
    </td>
  </tr>
  <tr>
    <td>0.10 (<a href="/blog/0.10-release-notes.html">Release Note</a>)</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.10.tar.gz
      @dl macOS darwin-amd64/elvish-v0.10.tar.gz
    </td>
    <td>
      @dl Linux linux-386/elvish-v0.10.tar.gz
    </td>
    <td>
      @dl Linux linux-arm64/elvish-v0.10.tar.gz
    </td>
  </tr>
  <tr>
    <td colspan="4" class="notice">
      Versions before 0.10 require cgo
    </td>
  </tr>
  <tr>
    <td>0.9 (<a href="/blog/0.9-release-notes.html">Release Note</a>)</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.9.tar.gz
      @dl macOS darwin-amd64/elvish-v0.9.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.8 (<a href="https://github.com/elves/elvish/releases/tag/v0.8">Release Note</a>)</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.8.tar.gz
      @dl macOS darwin-amd64/elvish-v0.8.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.7 (<a href="https://github.com/elves/elvish/releases/tag/v0.7">Release Note</a>)</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.7.tar.gz
      @dl macOS darwin-amd64/elvish-v0.7.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.6 (<a href="https://github.com/elves/elvish/releases/tag/v0.6">Release Note</a>)</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.6.tar.gz
      @dl macOS darwin-amd64/elvish-v0.6.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.5 (<a href="https://github.com/elves/elvish/releases/tag/v0.5">Release Note</a>)</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.5.tar.gz
      @dl macOS darwin-amd64/elvish-v0.5.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.4</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.4.tar.gz
      @dl macOS darwin-amd64/elvish-v0.4.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td colspan="4" class="notice">
      Versions before 0.4 do not use vendoring and cannot be reproduced
    </td>
  </tr>
  <tr>
    <td>0.3</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.3.tar.gz
      @dl macOS darwin-amd64/elvish-v0.3.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.2</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.2.tar.gz
      @dl macOS darwin-amd64/elvish-v0.2.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.1</td>
    <td>
      @dl Linux linux-amd64/elvish-v0.1.tar.gz
      @dl macOS darwin-amd64/elvish-v0.1.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
</table>
