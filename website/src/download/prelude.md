**Elvish is pre-release software**.

All binaries are statically linked.

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
</style>

<table>
  <colgroup>
    <col style="width:34%">
    <col style="width:22%">
    <col style="width:22%">
    <col style="width:22%">
  </colgroup>
  <tr>
    <th>Version</th>
    <th>Linux</th>
    <th>macOS</th>
    <th>Windows</th>
  </tr>
  <tr>
    <td>HEAD</td>
    <td>
      $dl amd64 elvish-linux-amd64-HEAD.tar.gz
      $dl 386 elvish-linux-386-HEAD.tar.gz
      $dl arm64 elvish-linux-arm64-HEAD.tar.gz
    </td>
    <td>
      $dl amd64 elvish-darwin-amd64-HEAD.tar.gz
    </td>
    <td>
      $dl amd64 elvish-windows-amd64-HEAD.zip
      $dl 386 elvish-windows-386-HEAD.zip
    </td>
  </tr>
  <tr>
    <td>
      0.11 (<a href="/blog/0.11-release-notes.html">Release Note</a>)
    </td>
    <td>
      $dl amd64 elvish-linux-amd64-0.11.tar.gz
      $dl 386 elvish-linux-386-0.11.tar.gz
      $dl arm64 elvish-linux-arm64-0.11.tar.gz
    </td>
    <td>
      $dl amd64 elvish-darwin-amd64-0.11.tar.gz
    </td>
    <td>
      $dl amd64 elvish-windows-amd64-0.11.zip
      $dl 386 elvish-windows-386-0.11.zip
    </td>
  </tr>
</table>

For users in China, the [mirror](https://mirrors.tuna.tsinghua.edu.cn/elvish)
hosted by TUNA may be faster.

If your environment is not listed above, you may still be able to build Elvish
from [source](https://github.com/elves/elvish).

Historical versions:

<table>
  <colgroup>
    <col style="width:34%">
    <col style="width:22%">
    <col style="width:22%">
    <col style="width:22%">
  </colgroup>
  <tr>
    <th>Version</th>
    <th>Linux</th>
    <th>macOS</th>
    <th>Windows</th>
  </tr>
  <tr>
    <td>0.10.1 (<a href="/blog/0.10-release-notes.html">Release Note</a>)</td>
    <td>
      $dl amd64 elvish-0.10.1-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.10.1-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.9 (<a href="/blog/0.9-release-notes.html">Release Note</a>)</td>
    <td>
      $dl amd64 elvish-0.9-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.9-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.8 (<a href="https://github.com/elves/elvish/releases/tag/0.8">Release Note</a>)</td>
    <td>
      $dl amd64 elvish-0.8-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.8-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.7 (<a href="https://github.com/elves/elvish/releases/tag/0.7">Release Note</a>)</td>
    <td>
      $dl amd64 elvish-0.7-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.7-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.6 (<a href="https://github.com/elves/elvish/releases/tag/0.6">Release Note</a>)</td>
    <td>
      $dl amd64 elvish-0.6-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.6-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.5 (<a href="https://github.com/elves/elvish/releases/tag/0.5">Release Note</a>)</td>
    <td>
    $dl amd64 elvish-0.5-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.5-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.4</td>
    <td>
      $dl amd64 elvish-0.4-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.4-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.3</td>
    <td>
      $dl amd64 elvish-0.3-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.3-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.2</td>
    <td>
      $dl amd64 elvish-0.2-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.2-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>0.1</td>
    <td>
      $dl amd64 elvish-0.1-linux.tar.gz
    </td>
    <td>
      $dl amd64 elvish-0.1-osx.tar.gz
    </td>
    <td>N/A</td>
  </tr>
</table>


# OS-Specific Packages

## Fedora

RPM Package for Fedora is available from [the FZUG Repo](https://github.com/FZUG/repo/wiki/Add-FZUG-Repository):

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
