Prebuilt, statically linked binaries for some common platforms are provided
below.
If your environment is not listed above, you may still be able to build Elvish
from [source](https://github.com/elves/elvish).
For users in China, the
[mirror](https://mirrors.tuna.tsinghua.edu.cn/elvish) hosted by TUNA may be
faster.

Note that Elvish is **pre-release software**. It can be unstable, and does not
maintain backward compatibility from version to version.

<style>
  table {
    border-collapse: collpase;
    width: 100%;
    margin-bottom: 16px;
  }
  td, th {
    border: 1px solid #aaa;
    text-align: left;
    padding: 0.4em;
  }
  tr:nth-child(even) {
    background-color: #ddd;
  }
  .notice {
    text-align: center;
  }
</style>

<table>
  <tr>
    <th>Version</th>
    <th>x86-64</th>
    <th>x86</th>
    <th>ARMv8</th>
  </tr>
  <tr>
    <td>HEAD</td>
    <td>
      $dl Linux linux-amd64/elvish-HEAD.tar.gz
      $dl macOS darwin-amd64/elvish-HEAD.tar.gz
      $dl FreeBSD freebsd-amd64/elvish-HEAD.tar.gz
      $dl NetBSD netbsd-amd64/elvish-HEAD.tar.gz
      $dl OpenBSD openbsd-amd64/elvish-HEAD.tar.gz
      $dl Windows windows-amd64/elvish-HEAD.zip
    </td>
    <td>
      $dl Linux linux-386/elvish-HEAD.tar.gz
      $dl Windows windows-386/elvish-HEAD.zip
    </td>
    <td>
      $dl Linux linux-arm64/elvish-HEAD.tar.gz
    </td>
  </tr>
  <tr>
    <td>
      0.12 (<a href="/blog/0.12-release-notes.html">Release Note</a>)
    </td>
    <td>
      $dl Linux linux-amd64/elvish-v0.12.tar.gz
      $dl macOS darwin-amd64/elvish-v0.12.tar.gz
      $dl Windows windows-amd64/elvish-v0.12.zip
    </td>
    <td>
      $dl Linux linux-386/elvish-v0.12.tar.gz
      $dl Windows windows-386/elvish-v0.12.zip
    </td>
    <td>
      $dl Linux linux-arm64/elvish-v0.12.tar.gz
    </td>
  </tr>
</table>


# OS-Specific Packages

## Arch Linux

Elvish PKGBUILDs are available in AUR. Install [`elvish`](https://aur.archlinux.org/packages/elvish/) (latest version) or [`elvish-git`](https://aur.archlinux.org/packages/elvish-git/) (HEAD) using your favorite AUR helper.

Alternatively, prebuilt packages can be obtained from [Arch Linux CN repository](https://www.archlinuxcn.org/archlinux-cn-repo-and-mirror/):

```elvish
# Add archlinuxcn repository
printf "[archlinuxcn]\nServer = http://repo.archlinuxcn.org/$arch" | sudo tee -a /etc/pacman.conf
# Install keyring
pacman -Sy archlinuxcn-keyring
pacman -S elvish
```

## Fedora

RPM packages are available from [the FZUG Repo](https://github.com/FZUG/repo/wiki/Add-FZUG-Repository):

```elvish
# Add FZUG repo
dnf config-manager --add-repo=http://repo.fdzh.org/FZUG/FZUG.repo
# Install Elvish
dnf install elvish
```

## Debian / Ubuntu

Elvish is packaged by [Debian](https://packages.debian.org/elvish) since
buster and by [Ubuntu](http://packages.ubuntu.com/elvish) since 17.10.

However, packages in official repositories are likely outdated. You can
install the latest release from
[PPA](https://launchpad.net/~zhsj/+archive/ubuntu/elvish):

```elvish
# Add Elvish PPA repo
sudo wget -O /etc/apt/trusted.gpg.d/elvish \
  'https://sks-keyservers.net/pks/lookup?search=0xE9EA75D542E35A20&options=mr&op=get'
sudo gpg --dearmor /etc/apt/trusted.gpg.d/elvish
sudo rm /etc/apt/trusted.gpg.d/elvish
echo 'deb http://ppa.launchpad.net/zhsj/elvish/ubuntu xenial main' |
  sudo tee /etc/apt/sources.list.d/elvish.list
sudo apt-get update

# Install Elvish
sudo apt-get install elvish
```

## macOS (Homebrew)

Elvish is packaged in Homebrew:

```elvish
# Install latest release
brew install elvish
# Or install HEAD:
brew install --HEAD elvish
```

## OpenBSD

Elvish is available in the official OpenBSD package repository. This will
install the latest release:

```elvish
doas pkg_add elvish
```

## NixOS (nix)

Elvish is packaged in [nixpkgs](https://github.com/NixOS/nixpkgs/blob/master/pkgs/shells/elvish/default.nix):

```elvish
# Install latest release
nix-env -i elvish
```

# Old versions

The following old versions are no longer supported. They are only listed
here for historical interest.

<table>
  <tr>
    <th>Version</th>
    <th>x86-64</th>
    <th>x86</th>
    <th>ARMv8</th>
  </tr>
  <tr>
    <td>
      0.11 (<a href="/blog/0.11-release-notes.html">Release Note</a>)
    </td>
    <td>
      $dl Linux linux-amd64/elvish-v0.11.tar.gz
      $dl macOS darwin-amd64/elvish-v0.11.tar.gz
      $dl Windows windows-amd64/elvish-v0.11.zip
    </td>
    <td>
      $dl Linux linux-386/elvish-v0.11.tar.gz
      $dl Windows windows-386/elvish-v0.11.zip
    </td>
    <td>
      $dl Linux linux-arm64/elvish-v0.11.tar.gz
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
      $dl Linux linux-amd64/elvish-v0.10.tar.gz
      $dl macOS darwin-amd64/elvish-v0.10.tar.gz
    </td>
    <td>
      $dl Linux linux-386/elvish-v0.10.tar.gz
    </td>
    <td>
      $dl Linux linux-arm64/elvish-v0.10.tar.gz
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
      $dl Linux linux-amd64/elvish-v0.9.tar.gz
      $dl macOS darwin-amd64/elvish-v0.9.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.8 (<a href="https://github.com/elves/elvish/releases/tag/v0.8">Release Note</a>)</td>
    <td>
      $dl Linux linux-amd64/elvish-v0.8.tar.gz
      $dl macOS darwin-amd64/elvish-v0.8.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.7 (<a href="https://github.com/elves/elvish/releases/tag/v0.7">Release Note</a>)</td>
    <td>
      $dl Linux linux-amd64/elvish-v0.7.tar.gz
      $dl macOS darwin-amd64/elvish-v0.7.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.6 (<a href="https://github.com/elves/elvish/releases/tag/v0.6">Release Note</a>)</td>
    <td>
      $dl Linux linux-amd64/elvish-v0.6.tar.gz
      $dl macOS darwin-amd64/elvish-v0.6.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.5 (<a href="https://github.com/elves/elvish/releases/tag/v0.5">Release Note</a>)</td>
    <td>
      $dl Linux linux-amd64/elvish-v0.5.tar.gz
      $dl macOS darwin-amd64/elvish-v0.5.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.4</td>
    <td>
      $dl Linux linux-amd64/elvish-v0.4.tar.gz
      $dl macOS darwin-amd64/elvish-v0.4.tar.gz
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
      $dl Linux linux-amd64/elvish-v0.3.tar.gz
      $dl macOS darwin-amd64/elvish-v0.3.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.2</td>
    <td>
      $dl Linux linux-amd64/elvish-v0.2.tar.gz
      $dl macOS darwin-amd64/elvish-v0.2.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.1</td>
    <td>
      $dl Linux linux-amd64/elvish-v0.1.tar.gz
      $dl macOS darwin-amd64/elvish-v0.1.tar.gz
    </td>
    <td>N/A</td>
    <td>N/A</td>
  </tr>
</table>
