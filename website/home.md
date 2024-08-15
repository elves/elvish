<div class="intro">
<div class="intro-content">

**Elvish** (*noun*):

1.  A powerful scripting language.

2.  A shell with useful interactive features built-in.

3.  A statically linked binary for Linux, BSDs, macOS or Windows.

</div>
<div class="action">
  <a href="get/" class="primary button">Download</a>
  <a href="learn/" class="button">Learn</a>
  <a href="https://github.com/elves/elvish" class="button" target="_blank">GitHub</a>
  <a href="#community" class="community button">Community</a>
  <a href="sponsor/" class="sponsor button">Sponsor</a>
</div>
</div>

<section>
<header>

Powerful modern shell scripting

</header>
<div class="showcase content">

Write readable and maintainable scripts - no cryptic operators, no
double-quoting every variable.

```elvish jpg-to-png.elv [(explainer)](learn/scripting-case-studies.html#jpg-to-png.elv)
for x [*.jpg] {
  gm convert $x (str:trim-suffix $x .jpg).png
}
```

Power up your workflows with data structures and functional programming.

```elvish update-servers-in-parallel.elv [(explainer)](learn/scripting-case-studies.html#update-servers-in-parallel.elv)
var hosts = [[&name=a &cmd='apt update']
             [&name=b &cmd='pacman -Syu']]
# peach = "parallel each"
peach {|h| ssh root@$h[name] $h[cmd] } $hosts
```

Catch errors before code executes.

```elvish-transcript Terminal: elvish [(explainer)](learn/scripting-case-studies.html#catching-errors-early)
~> var project = ~/project
~> rm -rf $projetc/bin
compilation error: variable $projetc not found
```

Command failures abort execution by default. No more silent failures, no more
`&&` everywhere.

```elvish-transcript Terminal: elvish [(explainer)](learn/scripting-case-studies.html#command-failures)
~> gm convert a.jpg a.png; rm a.jpg
gm convert: Failed to convert a.jpg
Exception: gm exited with 1
  [tty 1]:1:1-22: gm convert a.jpg a.png; rm a.jpg
# "rm a.jpg" is NOT executed
```

</div>
</section>
<section>
<header>

Run it anywhere

</header>
<div class="showcase content">

Elvish comes in a single statically linked binary for your laptop, your server,
your PC, or your Raspberry Pi.

```elvish-transcript Terminal: Raspberry Pi
~> wget dl.elv.sh/linux-arm64/elvish-HEAD.tar.gz
~> tar -C /usr/local/bin -xvf elvish-HEAD.tar.gz
elvish
~> elvish
```

Use Elvish in your CI/CD pipelines. Convenient shell syntax and modern
programming language - why not both?

```yaml github-actions.yaml
steps:
  - uses: elves/setup-elvish@v1
    with:
      elvish-version: HEAD
  - name: Run something with Elvish
    shell: elvish {0}
    run: |
      echo Running Elvish $version
```

</div>
</section>
<section>
<header>

Interactive shell with batteries included

</header>
<div class="showcase content">

Press <kbd>Ctrl-L</kbd> for directory history, and let Elvish find
`java/com/acme/project` for you.

```ttyshot Terminal: elvish - directory history [(more)](learn/tour.html#directory-history)
home/dir-history
```

Press <kbd>Ctrl-R</kbd> for command history. That beautiful `ffmpeg` command you
crafted two months ago is still there.

```ttyshot Terminal: elvish - command history [(more)](learn/tour.html#command-history)
home/cmd-history
```

Press <kbd>Ctrl-N</kbd> for the builtin file manager. Explore directories and
files without leaving the comfort of your shell.

```ttyshot Terminal: elvish - file manager [(more)](learn/tour.html#navigation-mode)
home/file-manager
```

</div>
</section>
<section>
<div class="columns content">
<div class="column">
<header id="community">

Talk with the community

</header>

-   Join the [Forum](https://bbs.elv.sh) to ask questions, share your
    experience, and show off your projects!

-   Join the chatroom to talk to fellow users in real time! The following
    channels are all bridged together thanks to [Matrix](https://matrix.org):

    -   Telegram: [Elvish user group](https://t.me/+Pv5ZYgTXD-YaKwcP)

    -   Discord: [Elvish Shell](https://discord.gg/jrmuzRBU8D)

    -   Matrix: [#users:elv.sh](https://matrix.to/#/#users:elv.sh)

    -   IRC: [#elvish](https://web.libera.chat/#elvish) on Libera Chat

    -   Gitter: [elves/elvish](https://gitter.im/elves/elvish)

</div>
<div class="column">
<header>

More resources

</header>

-   [Try Elvish](https://try.elv.sh) directly from the browser (beta)

-   [Awesome Elvish](https://github.com/elves/awesome-elvish): Official list of
    unofficial Elvish modules

-   [@ElvishShell](https://twitter.com/elvishshell) on Twitter

-   [Elvish TV](https://www.youtube.com/@ElvishShell) on YouTube

</div>
</div>
</section>
