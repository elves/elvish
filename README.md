An experimental Unix shell
==========================

This is a work in progress. Things may change and/or break without notice. You
have been warned...

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
