Welcome to the second issue of Elvish Newsletter!

Elvish is a shell that seeks to combine a full-fledged programming language with
a friendly user interface. This newsletter is a summary of its progress and
future plans.

# Release of 0.10 Version

This newsletter accompanies the release of the 0.10 version. This release
contains 125 commits, with contributions from @xofyarg, @tw4452852, @ALSchwalm,
@zhsj, @HeavyHorst, @silvasur, @zzamboni, @Chilledheart, @myfreeweb, @xchenan
and @jiujieti.

Elvish used to depend on SQLite for storage. As a result, compiling Elvish
relied on cgo and required a C compiler. This release sees the switch to BoltDB,
making Elvish a pure-Go project. Elvish can now be compiled much faster, and
into a fully statically linked binary. Cross-compilation is also much easier, as
the Go compiler has fantastic cross-compiling support.

Maps (`[&k=v &k2=v2]`) are now implemented using persistent hash maps. This
concludes the transition to persistent data structures for all primary data
types (strings, lists, maps). Persistent data structures are immutable, and thus
have a simpler semantics and are automatically concurrency-safe. This does have
an interesting impact on the semantics of assignments, which is now documented
in a new section on the [unique semantics](../learn/unique-semantics.html) page.

For a complete list of changes, see the
[release notes](0.10-release-notes.html).

# Community

-   We now have an official list of awesome unofficial Elvish libraries:
    [elves/awesome-elvish](https://github.com/elves/awesome-elvish). Among
    others, we now have at least two very advanced prompt themes, chain.elv from
    @zzamboni and powerline.elv from @muesli :)

-   Diego Zamboni (@zzamboni), the author of chain.elv, has written very
    passionately on Elvish:
    [Elvish, an awesome Unix shell](http://zzamboni.org/post/elvish-an-awesome-unix-shell/).

-   Patrick Callahan has given an awesome talk on
    [Delightful Command-Line Experiences](https://dl.elvish.io/resources/callahan-delightful-commandline-experiences.pdf),
    featuring Elvish as a "very lively, ambitious shell".

-   Elvish is now [packaged](https://packages.debian.org/elvish) in Debian.

-   The number of followers to
    [@RealElvishShell](https://twitter.com/RealElvishShell/) has grown to 23.

# Plans

The mid-term remains the same as in the
[previous issue](newsletter-july-2017.html): stabilizing the language core and
enhancing usability of the user interface.

The short-term plan is captured in the
[milestone](https://github.com/elves/elvish/milestone/3) for the 0.11 version.
Among other things, 0.11 is expected to ship with `epm`,
[the standard package manager](https://github.com/elves/elvish/issues/239) for
Elvish, and a more responsive interface by running
[prompts](https://github.com/elves/elvish/issues/482) and
[completions](https://github.com/elves/elvish/issues/483) asynchronously. Stay
very tuned.

# Conclusions

In the last newsletter, I predicted that we will be featuring *Elvish for Python
Users* and *Tetris in Your Shell* in a future newsletter. It seems we are
getting close to that pretty steadily.

Have fun with Elvish!

\- xiaq
