# A novel Unix shell

[![GoDoc](http://godoc.org/github.com/elves/elvish?status.svg)](http://godoc.org/github.com/elves/elvish)
[![Build Status on Travis](https://travis-ci.org/elves/elvish.svg?branch=master)](https://travis-ci.org/elves/elvish)

This project aims to explore the potentials of the Unix shell. It is a work in
progress; things will change without warning. The [issues list](https://github.com/elves/elvish/issues) contains many things I'm working on.

## Screenshot

Elvish looks like this:

![syntax highlighting](https://raw.githubusercontent.com/elves/images/master/syntax.png)

## Prebuilt binaries

64-bit Linux: `curl -s https://dl.elvish.io/elvish-linux.tar.gz | sudo tar vxz -C /usr/bin`

64-bit Mac OS X: `curl -s https://dl.elvish.io/elvish-osx.tar.gz | sudo tar vxz -C /usr/bin`

See also [Building Elvish](#building-elvish).

## Getting Started

Elvish mimics bash and zsh in a lot of places. The following shows some key differences and highlights, as well as some common tasks:

* Put your startup script in `~/.elvish/rc.elv`. There is no `alias` yet, but you can achieve the goal by defining a function:

  ```sh
  fn ls { external:ls --color $@ }
  ```

  The `external:` prefix ensures that the external command named `ls` will be called. Otherwise this definition will result in infinite recursion.

* The left and right prompts and be customized by modifying `le:prompt` and `le:rprompt`. They can be assigned either to a function, in which their outputs are used, or a constant string. The following simulates the default prompts but uses fancy Unicode:

  ```sh
  # Changes during a session; use function.
  # "tilde-abbr" abbreviates home directory to a tilde.
  le:prompt={ put `tilde-abbr $pwd`'❱ ' }
  # Doesn't change during a session; use constant string.
  le:rprompt=`whoami`✸`hostname`
  ```

* Press Up to search through history. It uses what you have typed to do prefix match. To cancel, press Escape.

  ![history](https://raw.githubusercontent.com/elves/images/master/history.png)

* Press Tab to start completion. Use arrow key and Tab to select the candidate;  press Enter, or just continue typing to accept. To cancel, press Escape.

  ![tab completion](https://raw.githubusercontent.com/elves/images/master/completion.png)

* Press Ctrl-N to start navigation mode. Press Ctrl-H to show hidden files; press again to hide. Press tab to append selected filename to your command. Likewise, pressing Escape gets you back to the default (insert) mode.

  ![navigation mode](https://raw.githubusercontent.com/elves/images/master/navigation.png)

* Try typing `echo [` and press Enter. Elvish knows that the command is unfinished due to the unclosed `[` and inserts a newline instead of accepting the command. Moreover, common errors like syntax errors and missing variables are highlighted in real time.

* Elvish remembers which directories you have visited. Press Ctrl-L to list visited directories. Like in completion, use Up, Down and Tab to navigate and use Enter to accept (which `cd`s into the selected directory). Press Escape to cancel.

  ![location mode](https://raw.githubusercontent.com/elves/images/master/location.png)

  Type to filter:
  
  ![location mode, filtering](https://raw.githubusercontent.com/elves/images/master/location-filter.png)

  The filtering algorithm takes your filter and adds `**` to both sides of each path component. So `g/di` becomes pattern `**g**/**di**`, so it matches /home/xiaq/**g**o/elves/elvish/e**di**t.

* **NOTE**: Default key bindings as listed above are subject to change in the future; but the functionality will not go away.

* Elvish doesn't support history expansion like `!!`. Instead, it has a "bang mode", triggered by `Alt-,`, that provides the same functionality. For example, if you typed a command but forgot to add `sudo`, you can then type `sudo ` and press `Alt-,` twice to fix it:

  ![bang mode](https://raw.githubusercontent.com/elves/images/master/bang.png)

* Lists look like `[a b c]`, and maps look like `[&key1=value1 &key2=value2]`. Unlike other shells, a list never expands to multiple words, unless you explicitly splice it by prefixing the variable name with `$@`:
  ```sh
  ~> li=[1 2 3]
  ~> for x in $li; do echo $x; done
  [1 2 3]
  ~> for x in $@li; do echo $x; done
  1
  2
  3
  ```

* You can manipulate search paths through the special list `$paths`:
  ```sh
  ~> echo $paths
  [/bin /sbin]
  ~> paths=[/opt/bin $@paths /usr/bin]
  ~> echo $paths
  [/opt/bin /bin /sbin /usr/bin]
  ~> echo $env:PATH
  /opt/bin:/bin:/sbin:/usr/bin
  ```

* You can manipulate the keybinding through the map `$le:binding`. For example, this binds Ctrl-L to clearing the terminal: `le:binding[insert][Ctrl-L]={ clear > /dev/tty }`. The first index is the mode and the second is the key. (Yes, the braces enclose a lambda.)

  Use `pprint $le:binding` to get a nice (albeit long) view of the current keybinding.

* Environment variables live in a separate `env:` namespace and must be explicitly qualified:
  ```sh
  ~> put $env:HOME
  ▶ /home/xiaq
  ~> env:PATH=$env:PATH":/bin"
  ```

* There is no interpolation inside double quotes (yet). Use implicit string concatenation:
  ```sh
  ~> name=xiaq
  ~> echo "My name is "$name"."
  My name is xiaq.
  ```

* Elementary floating-point arithmetics as well as comparisons are builtin. Unfortunately, you have to use prefix notation:
  ```sh
  ~> + 1 2
  ▶ 3
  ~> / `* 2 3` 4
  ▶ 1.5
  ~> / (* 2 3) 4 # parentheses are equivalent to backquotes, but look nicer in arithmetics
  ▶ 1.5
  ~> > 1 2 # ">" may be used as a command name
  false
  ~> < 1 2 # "<" may also be used as a command name; silence means "true"
  ```

* Functions are defined with `fn`. You can name arguments:
  ```sh
  ~> fn square [x]{
       * $x $x
     }
  ~> square 4
  ▶ 16
  ```

* Output of some builtin commands start with a funny "▶". It is not part of the output itself, but shows that such commands output a stream of values instead of bytes. As such, their internal structures as well as boundaries between values are preserved. This allows us to manipulate structured data in the shell; more on this later.


## Building Elvish

Go >= 1.5 is required. Linux is fully supported. It is likely to work on BSDs and Mac OS X. Windows is **not** supported yet.

The main binary can be installed using `go get github.com/elves/elvish`. There is also an auxiliary program called elvish-stub; install it with `make stub`. Elvish is functional without the stub, but job control features depend on it.

If you are lazy and use `bash` or `zsh` now, here is something you can copy-paste into your terminal:

```sh
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
mkdir -p $GOPATH

go get github.com/elves/elvish
make -C $GOPATH/src/github.com/elves/elvish stub

for f in ~/.bashrc ~/.zshrc; do
    printf 'export %s=%s\n' GOPATH '$HOME/go' PATH '$PATH:$GOPATH/bin' >> $f
done
```

[How To Write Go Code](http://golang.org/doc/code.html) explains how `$GOPATH` works.


## Name

In [roguelikes](https://en.wikipedia.org/wiki/Roguelike), items made by the elves have a reputation of high quality.  These are usually called **elven** items, but I chose **elvish** for an obvious reason.

The adjective for elvish is also "elvish", not "elvishy" and definitely not "elvishish".
