<!-- toc -->

# Why a new shell?

The author of Elvish found the concept of a shell -- a programmable, interactive
text interface for an OS -- to be simple yet powerful. However, he also felt
that the more traditional shells, such as `bash` and `zsh`, didn't take it far
enough. They all had *some* programming capabilities, but they were only really
sufficient for manipulating strings or lists of strings. They also had *some*
nice interactive features, such as tab completion, but more advanced UI features
were either nonexistent, hidden behind obscure configuration options, or
required external programs with suboptimal integration with the rest of the
shell experience.

So the author set out to build a shell that he personally wished existed. This
was done by starting from the basic concept of a shell -- a programmable,
interactive text interface for an OS -- and rethinking how it could be executed
with modern techniques. Elvish is that product, and the rethinking is still
ongoing.

There are a lot of other projects with similar motivations;
[this list](https://github.com/oilshell/oil/wiki/Alternative-Shells) on Oil's
wiki offers a quick overview.

# Can I use Elvish as my shell?

Yes. Many people already use Elvish as their daily drivers. Follow the
instructions in the [Get Elvish](../get/) page to install Elvish and use it as
your default shell.

# Why is Elvish incompatible with bash?

The author felt it would be easier and more fun to start with a clean slate.

However, it is certainly possible to build a powerful new shell that is also
compatible; Elvish just happens to not be that. If you definitely need
compatibility with traditional shells, [Oil](http://www.oilshell.org/) is a
promising project.

# Why is Elvish restricted to terminals?

There are a lot of things to dislike about VT100-like terminals, but they are
still the most portable API for building a text interface. All major desktop
operating systems (including
[Windows](https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences))
support this API, and many other programs target this API.

That said, the few parts of Elvish that rely on a VT100-like terminal is only
loosely coupled with the rest of Elvish. It will be very easy to port Elvish to
other interfaces, and the author might explore this space at a later stage.

# Why is Elvish called Elvish?

Elvish is named after **elven** items in
[roguelikes](https://en.wikipedia.org/wiki/Roguelike), which has a reputation of
high quality. You can think of Elvish as an abbreviation of "elven shell".

The name is not directly related to
[Tolkien's Elvish languages](https://en.wikipedia.org/wiki/Elvish_languages_(Middle-earth)),
but you're welcome to create something related to both Elvishes.

Alternatively, Elvish is a backronym for "Expressive programming Language and
Versatile Interactive SHell".
