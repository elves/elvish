<!-- toc -->

Installing Elvish with a package manager allows you to upgrade Elvish alongside
the rest of your system, but may not give you the latest version.

If the package manager you use doesn't have the latest version, you are strongly
recommended to the [official binary](./) instead. For a comprehensive list of
packages and their freshness, see
[this Repology page](https://repology.org/project/elvish/versions).

# Arch Linux

![Arch package](https://repology.org/badge/version-for-repo/arch/elvish.svg)

To install the latest packaged release:

```elvish
pacman -S elvish
```

To install the HEAD version, install
[`elvish-git`](https://aur.archlinux.org/packages/elvish-git/) from AUR with
your favorite AUR helper:

```elvish
yay -S elvish-git
```

# Debian / Ubuntu

![Debian 13 package](https://repology.org/badge/version-for-repo/debian_13/elvish.svg)
![Debian Unstable package](https://repology.org/badge/version-for-repo/debian_unstable/elvish.svg)
![Ubuntu 23.10 package](https://repology.org/badge/version-for-repo/ubuntu_23_10/elvish.svg)
![Ubuntu 24.04 package](https://repology.org/badge/version-for-repo/ubuntu_24_04/elvish.svg)

Elvish is packaged by [Debian](https://packages.debian.org/elvish) since buster
and by [Ubuntu](http://packages.ubuntu.com/elvish) since 17.10:

```elvish
apt install elvish
```

# Fedora

![Fedora 40 package](https://repology.org/badge/version-for-repo/fedora_40/elvish.svg)
![Fedora Rawhide package](https://repology.org/badge/version-for-repo/fedora_rawhide/elvish.svg)

Elvish is packaged for [Fedora](https://packages.fedoraproject.org/pkgs/elvish).
To install it with `dnf`:

```elvish
dnf install elvish
```

# macOS

Elvish is packaged by both [Homebrew](https://brew.sh) and
[MacPorts](https://www.macports.org).

![Homebrew package](https://repology.org/badge/version-for-repo/homebrew/elvish.svg)

To install from Homebrew:

```elvish
# Install latest packaged release
brew install elvish
# Or install HEAD:
brew install --HEAD elvish
```

![MacPorts package](https://repology.org/badge/version-for-repo/macports/elvish.svg)

To install from MacPorts:

```elvish
sudo port selfupdate
sudo port install elvish
```

# Windows

![Scoop package](https://repology.org/badge/version-for-repo/scoop/elvish.svg)

Elvish is available in the Main
[bucket](https://github.com/ScoopInstaller/Main/blob/master/bucket/elvish.json)
of [Scoop](https://scoop.sh). This will install the latest packaged release:

```elvish
scoop install elvish
```

# FreeBSD

![FreeBSD port](https://repology.org/badge/version-for-repo/freebsd/elvish.svg)

Elvish is available in the FreeBSD ports tree and as a prebuilt package. Both
methods will install the latest packaged release.

To install with `pkg`:

```elvish
pkg install elvish
```

To build from the ports tree:

```elvish
cd /usr/ports/shells/elvish
make install
```

# NetBSD / pkgsrc

![pkgsrc current package](https://repology.org/badge/version-for-repo/pkgsrc_current/elvish.svg)

Elvish is [available in pkgsrc](https://pkgsrc.se/shells/elvish). To install
from a binary package, run the following command:

```elvish
pkgin install elvish
```

To build the elvish package from source instead:

```elvish
cd /usr/pkgsrc/shells/elvish
make package-install
```

# OpenBSD

![OpenBSD port](https://repology.org/badge/version-for-repo/openbsd/elvish.svg)

Elvish is available in the official OpenBSD package repository. This will
install the latest packaged release:

```elvish
doas pkg_add elvish
```

# NixOS (nix)

![nixpkgs stable 23.11 package](https://repology.org/badge/version-for-repo/nix_stable_23_11/elvish.svg)
![nixpkgs unstable package](https://repology.org/badge/version-for-repo/nix_unstable/elvish.svg)

Elvish is packaged in
[nixpkgs](https://github.com/NixOS/nixpkgs/blob/master/pkgs/shells/elvish/default.nix):

```elvish
# Install latest packaged release
nix-env -i elvish
```
