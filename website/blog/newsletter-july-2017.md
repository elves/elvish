Welcome to the first issue of Elvish Newsletter!

Elvish is a shell that seeks to combine a full-fledged programming language with
a friendly user interface. This newsletter is a summary of its progress and
future plans.

# Status Updates

-   18 pull requests to the [main repo](https://github.com/elves/elvish) have
    been merged in the past four weeks. Among them 13 were made by @xofyargs,
    and the rest by @myfreeweb, @jiujieti, @HeavyHorst, @silvasur and
    @ALSchwalm. The [website repo](https://github.com/elves/elvish.io) has also
    merged 3 pull requests from @bengesoff, @zhsj and @silvasur. Many kudos!

-   The [website](https://elvish.io) was [officially live](../blog/live.html) on
    3 July. Although the initial
    [submission](https://news.ycombinator.com/item?id=14691639) to HN was a
    failure, Elvish gained
    [quite](https://www.reddit.com/r/programming/comments/6l38nd/elvish_friendly_and_expressive_shell/)
    [some](https://www.reddit.com/r/golang/comments/6l3aev/elvish_friendly_and_expressive_shell_written_in_go/)
    [popularity](https://www.reddit.com/r/linux/comments/6l6wcs/elvish_friendly_and_expressive_shell_now_ready/)
    on Reddit, and [another](https://news.ycombinator.com/item?id=14698187) HN
    submission made to the homepage. These, among others, have brought 40k
    unique visitors to the website, totalling 340k HTTP requests. Thank you
    Internet :)

-   A lot of discussions have happened over the IM channels and the issue
    tracker, and it has become necessary to better document the current status
    of Elvish and organize the development effort, and this newsletter is part
    of the response.

    There is no fixed schedule yet, but the current plan is to publish
    newsletters roughly every month. Preview releases of Elvish, which used to
    happen quite arbitrarily, will also be done to coincide with the publication
    of newsletters.

-   There are now IM channels for developers, see below for details.

# Short-Term and Mid-Term Plans

The next preview release will be 0.10, and there is now a
[milestone](https://github.com/elves/elvish/milestone/2) for it, a list of
issues considered vital for the release. If you would like to contribute, you
are more than welcome to pick an issue from that list, although you are also
more than welcome to pick just any issue.

Aside from the short-term goal of releasing 0.10, here are the current mid-term
focus areas of Elvish development:

-   Stabilizing the language core.

    The core of Elvish is still pretty immature, and it is definitely not as
    usable as any other dynamic language, say Python or Clojure. Among others,
    the 0.10 milestone now plans changes to the implementation of maps
    ([#414](https://github.com/elves/elvish/issues/414)), a new semantics of
    element assignment ([#422](https://github.com/elves/elvish/issues/422)) and
    enhanced syntax for function definition
    ([#82](https://github.com/elves/elvish/issues/82) and
    [#397](https://github.com/elves/elvish/issues/397)). You probably wouldn't
    expect such fundamental changes in a mature language :)

    A stable language core is a prerequisite for a 1.0 release. Elvish 1.x will
    maintain backwards compatibility with code written for earlier 1.x versions.

-   Enhance usability of the user interface, and provide basic programmability.

    The goal is to build a fully programmable user interface, and there are a
    lot to be done. Among others, the 0.10 milestone plans to support
    manipulating the cursor ([#415](https://github.com/elves/elvish/issues/415))
    programmatically, scrolling of previews in navigation mode previews
    ([#381](https://github.com/elves/elvish/issues/381)), and invoking external
    editors for editing code
    ([#393](https://github.com/elves/elvish/issues/393)).

    The user interface is important for two reasons. Enhancements to the UI can
    improve the power of Elvish directly and significantly; its API is also a
    very good place for testing the language. By developing the language and the
    user interface in parallel, we can make sure that they work well together.

Like many other open source projects, you are welcome to discuss and challenge
the current plan, or come up with your ideas regarding the design and
implementation.

(So what's the long-term goal of Elvish? The long-term goal is to remove the
"seeks to" part from the introduction of Elvish at the beginning of the post.)

# Development IM Channels

To better coordinate development, there are now IM channels for Elvish
development: [#elvish-dev](http://webchat.freenode.net/?channels=elvish-dev) on
freenode, [elves/elvish-dev](https://gitter.im/elves/elvish-dev) on Gitter and
[@elvish_dev](https://telegram.me/elvish_dev) on Telegram. These channels are
all connected together thanks to [fishroom](https://github.com/tuna/fishroom).

For general questions, you are welcome in
[#elvish](https://webchat.freenode.net/?channels=elvish) on Freenode,
[elves/elvish-public](https://gitter.im/elves/elvish-public) on Gitter, or
[@elvish](https://telegram.me/elvish) on Telegram.

# Conclusion

This concludes this first issue of the newsletter. Hopefully future issues of
this newsletter will also feature blog posts from Elvish users like *Elvish for
Python Users* and popular Elvish modules like *Tetris in Your Shell* :)

Have Fun with Elvish!

\- xiaq
