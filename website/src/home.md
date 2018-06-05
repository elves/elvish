**Elvish** is a friendly and expressive shell for Linux, BSDs, macOS and
Windows.

<!--
<pre id="demo-debug">
</pre>
-->

<ul id="demo-switcher">
  <li id="demo-js-warning">
    Enable JavaScript for the option of showing demos as slides.
  </li>
  <li id="demo-expander-li" class="no-display"><a id="demo-expander">↧</a></li>
</ul>

<div id="demo-window"> <div id="demo-container" class="expanded">
  <div class="demo-wrapper"> <div class="demo">
    <div class="demo-col left"><div class="demo-ttyshot">
      $ttyshot pipelines
    </div></div>
    <div class="demo-col right"> <div class="demo-description">
      <div class="demo-title">Powerful Pipelines</div>
      <p>
        Text pipelines are intuitive and powerful. However, if your data have
        inherently complex structures, passing them through the pipeline
        often requires a lot of ad-hoc, hard-to-maintain text processing code.
      </p>
      <p>
        Pipelines in Elvish can carry structured data, not just text. You can
        stream lists, maps and even functions through the pipeline.
      </p>
    </div> </div>
  </div> </div>

  <div class="demo-wrapper"> <div class="demo">
    <div class="demo-col left"><div class="demo-ttyshot">
      $ttyshot control-structures
    </div></div>
    <div class="demo-col right"> <div class="demo-description">
      <div class="demo-title">Intuitive Control Structures</div>
      <p>
        If you know programming, you probably already know how
        <code>if</code> looks in C. So why learn another syntax?
      </p>
      <p>
        Control structures in Elvish have an intuitive C-like syntax.
      </p>
    </div> </div>
  </div> </div>

  <div class="demo-wrapper"> <div class="demo">
    <div class="demo-col left"><div class="demo-ttyshot">
      $ttyshot location-mode
    </div></div>
    <div class="demo-col right"> <div class="demo-description">
      <div class="demo-title">Directory History</div>
      <p>
        Is <code>cd /a/long/nested/directory</code> the first thing you
        do every day? Struggling to remember where your logs and
        configurations?
      </p>
      <p>
        Elvish remembers where you have been. Press Ctrl-L and search, like in a
        browser.
      </p>
    </div> </div>
  </div> </div>

  <div class="demo-wrapper"> <div class="demo">
    <div class="demo-col left"><div class="demo-ttyshot">
      $ttyshot histlist-mode
    </div></div>
    <div class="demo-col right"> <div class="demo-description">
      <div class="demo-title">Command History</div>
      <p>
        Want to find the magical <code>ffmpeg</code> command that you used to
        transcode a video file two months ago?
      </p>
      <p>
        Just dig through your command history with Ctrl-R. Same key, more
        useful.
      </p>
      <p>
        (To be fair, you can do this in bash with <code>history | grep
        ffmpeg</code>, but it's far fewer keystrokes in Elvish :)
      </p>
    </div> </div>
  </div> </div>

  <div class="demo-wrapper"> <div class="demo">
    <div class="demo-col left"><div class="demo-ttyshot">
      $ttyshot navigation-mode
    </div></div>
    <div class="demo-col right"> <div class="demo-description">
      <div class="demo-title">Built-in File Manager</div>
      <p>
        Power of the shell or convenience of a file manager?
      </p>
      <p>
        Choose both. Press Ctrl-N to quickly navigate directories and preview
        files, with full shell power.
      </p>
    </div> </div>
  </div> </div>
</div> </div>

<link href="/assets/home-demos.css" rel="stylesheet">
<script src="/assets/home-demos.js"></script>

# Getting Elvish

Elvish is still in development, but has enough features and stability for
daily use.

*   [Download](/download/) prebuilt binaries if you are running Linux or macOS on
    an x86-64 CPU.

*   Source code is available on the [GitHub repository](https://github.com/elves/elvish).

# Speaking Elvish

*   [Learn](/learn/) to speak Elvish by following tutorials.

    If you are not experienced with any shell, start with the
    [fundamentals](/learn/fundamentals.html). (This tutorial is still a work in
    progress, though.)

    If you come from other shells, read the [cookbook](/learn/cookbook.html)
    to get started quickly, and learn about Elvish's [unique
    semantics](learn/unique-semantics.html).

*   [Reference](/ref/) documents describe Elvish in a more formal and complete
    way.

    Read about the [philosophy](/ref/philosophy.html), the
    [language](/ref/language.html), the [builtin module](/ref/builtin.html), and
    more.

*   [The blog](/blog/) contains news on Elvish.

    It is the place for release notes, notes on the internals of Elvish, and
    other announcements or musings from the developers.

*   [The feed](/feed.atom) contains updates to all sections of the website (not
    just the blog).

*   [Follow](https://twitter.com/RealElvishShell/) Elvish on Twitter.

# Meeting Other Elves

*   Join [#elvish](https://webchat.freenode.net/?channels=elvish) on Freenode,
    [elves/elvish-public](https://gitter.im/elves/elvish-public) on Gitter, or
    [@elvish](https://telegram.me/elvish) on Telegram.

    The wonderful [fishroom](https://github.com/tuna/fishroom) service
    connects all of them together. So just join whichever channel suits you
    best, and you won't miss discussions happening in other channels.

*   If you are interested in contributing to Elvish, you can also discuss at
    [#elvish-dev](http://webchat.freenode.net/?channels=elvish-dev) on
    freenode, [elves/elvish-dev](https://gitter.im/elves/elvish-dev) on Gitter
    or [@elvish_dev](https://telegram.me/elvish_dev) on Telegram.

*   Chinese speakers are also welcome in
    [#elvish-zh](https://webchat.freenode.net/?channels=elvish-zh) (Freenode)
    and [@elvish_zh](https://telegram.me/elvish_zh) (Telegram). There are
    also [#elvish-dev-zh](https://webchat.freenode.net/?channels=elvish-dev-zh)
    (Freenode) and [@elvish_dev_zh](https://telegram.me/elvish_dev_zh) (Telegram).

*   The [issue tracker](https://github.com/elves/elvish/issues) is the place
    for bug reports and feature requests.
