An experimental Unix shell
==========================

This is a work in progress. Things may change and/or break without notice. You
have been warned...

The Editor
----------

Those marked with ✔ are implemeneted (but could be broken from time to
time).

Like fish:

* Syntax highlighting ✔
* Auto-suggestion

Like zsh:

* Right-hand-side prompt ✔
* Dropdown menu completion ✔
* Programmable line editor

And:

* A vi keybinding that makes sense
* More intuive multiline editing
* An intuive method to save typed snippets into a script

The Language
------------

(Like the previous section, only those marked with ✔ have been implemented.)

* Running external programs and pipelines, of course: ✔
  ```
  > vim README.md
  ...
  > cat -v /dev/random
  ...
  > dmesg | grep bar
  ...
  ```

* Basically prefix syntax without the outmost pair of parentheses (`>`
  represents the prompt):
  ```
  > + 1 2
  3
  > * (+ 1 2) 3
  9
  ```

* Use backquote for literal string (so that you can write both single and
  double quotes inside), double backquotes for a literal backquote: ✔
  ```
  > echo `"He's dead, Jim."`
  "He's dead, Jim."
  > echo ```He's dead, Jim."`
  ``He's dead, Jim."
  ```

* Barewords are string literals:
  ```
  > = a `a`
  true
  ```

* Tables are a hybrid of array and hash (a la Lua); tables are first-class
  values: ✔
  ```
  > println [a b c &key value]
  [a b c &key value]
  > println [a b c &key value][0]
  a
  > println [a b c &key value][key]
  value
  ```

* Declare variable with `var`, set value with `set`; `var` also serve as a
  shorthand of var-set combo: ✔
  ```
  > var v
  > set v = [foo bar]
  > var u = [foo bar] # equivalent
  ```

* First-class closures, lisp-like functional programming:
  ```
  > map {|x| * 2 $x} [1 2 3]
  [2 4 6]
  > filter {|x| > $x 2} [1 2 3 4 5]
  [3 4 5]
  > map {|x| * 2 $x} (filter {|x| > $x 2} [1 2 3 4 5])
  ```

* Get rid of lots of irritating superfluous parentheses with pipelines (`put`
  is the builtin for outputting compound data):
  ```
  > put [1 2 3 4 5] | filter {|x| > $x 2} | map {|x| * 2 $x}
  [6 8 10]
  ```

* Use the table `$env` for environmental variables:
  ```
  > put $env[HOME]
  /home/xiaq
  > set env[PATH] = $env[PATH]:/bin
  ```

There are many parts of the language that is not yet decided. See TODO.md for
a list of things I'm currently thinking about.

Motivation
----------

The basic idea is to have a shell that is also a decant programming language.
Shells have long rejected such things as data structure beyond text arrays.
Some support associative arrays, which may be assigned to variables but not
passed as arguments to builtins or shell functions, making the use of them
very tedious.

The lesson of Tcl has taught us the "everything is (unstructured) text"
philosophy, which is also the idea underpinning classical Unix pipelines, is
too limited for proper programming. Indeed, the power of Tcl often lies in the
dynamic interpretation of text, assuming some predefined structure in it. Yet
with shells, where such facilities are basically nonexistent, it requires
great discipline to build maintainable software. Traditional initscripts,
program wrapper scripts and some of the more tricky tab-completion scripts are
notable examples.

However, the shell does come with a very powerful abstraction - the pipeline.
It is basically a facility for concatenative programming. Consider the
following code in lisp:

```
(set coll' (map f (filter pred coll)))
```

Written concatenatively, this can be - assuming `put` puts the argument to
output (akin to `echo`), and `set` can take data from input in place of in the
argument list:

```
put $coll | filter pred | map f | set coll2
```

The concatenative approach is much more natural (try reading both versions
aloud).

Another defining character of shells is the easiness to invoke external
programs; comparing `subprocess.call(['ls', '-l'])` with `ls -l` - the
difference is clear. Being easy to invoking external programs is what makes
shells shells *in its original sense*, i.e. user interface to the operating
system.

Putting together, the idea of this new Unix shell is starting from pipelines
and external program interaction, adding in programming-language-ish flavors,
towards building a decant programming language with a friendly (command line)
user interface, suitable for both *back-of-the-envolope* computation **and**
building more complex (but maybe not too complex!) software.

This is not exactly an ambitious goal, but it's something I have always
dreamed of.

Building
--------

You need go >= 1.1 to build this. Just run `make`. The resulting binary is
called `das`.

Name
----

Indeed, **das** is not a very good name for a Unix shell. The name is actually
a corrupted form of **dash**, which also happens to be the German definite
neuter article.

I have some other ideas in mind. One of them is **elv**, since I found
"elvish" to be a great adjective - I can't use "elf" though, since it's
already [taken](http://www.cs.cmu.edu/~fp/elf.html) and may be confused with
the well known [file
format](http://en.wikipedia.org/wiki/Executable_and_Linkable_Format).

Another possible source of names is the names of actual seashells; but my
English vocabulary is too small for me to recall any beyond "nautilus", which
is both too long and already taken.

I'm not avoiding names ending in "sh" though; but I do find "bash" to be a
terrible name. "fish" is clever, but it has a quite [unpleasant
adjective](https://en.wiktionary.org/wiki/fishy). I find "dash" really good
though, which is why it came to my mind :).

License
-------

BSD 2-clause license.  See LICENSE for a copy.
