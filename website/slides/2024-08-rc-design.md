# Designing a shell language for the 2010s

Qi Xiao (xiaq)

2024-08-01 @ Recurse Center

***

# Design principles

-   Starting point

    -   A modern general-purpose language

    -   ... inheriting the *spirit* and *feel* of traditional shells

    -   ⇒ A language that can scale...

        -   down, to one-liners

        -   up, to moderately complex projects

-   My design decisions

    -   *One* scalable language

        -   No separate features for interactive use vs scripting

        -   No separate command vs expression syntax

    -   Modern, but not experimental

        -   Use established ideas (hence the 2010s)

        -   But maybe combine them in new ways

***

# What's the spirit of the shell anyway?

-   The original scripting language

-   Old dichotomy

    -   System languages: compiled, statically typed, optimized for machine
        speed

    -   Scripting languages: interpreted, dynamically typed, optimized for human
        speed

-   Boundaries have blurred

-   A more essential identity

    -   Highly interactive: quick feedback loop, inspectable runtime

    -   Good at interfacing the outside world - programs written by other
        people, data from elsewhere

***

# What does the language look like?

-   ```elvish
    # jpg-to-png.elv
    for x [*.jpg] {
        gm convert $x (str:trim-suffix $x .jpg).png
    }
    ```

-   ```elvish
    # update-servers-in-parellel.elv
    var hosts = [[&name=a &cmd='apt update']
                 [&name=b &cmd='pacman -Syu']]
    # peach = "parallel each"
    peach {|h| ssh root@$h[name] $h[cmd] } $hosts
    ```

-   [Detailed explainers](https://elv.sh/learn/scripting-case-studies.html)

***

# Inheriting traditions

-   Barewords are strings

    ```elvish
    vim design.md
    echo $pid # Variables need $
    ```

-   Uniform, concise command syntax

    ```elvish
    cd d                # A builtin command
    vim design.md       # An external command
    if (eq a b) { ... } # A control-flow command
    ```

-   Information flows via input and output

    ```elvish
    cat *.go | wc -l       # Either via pipeline
    mkdir (date +%Y-%m-%d) # Or via output capture
    ```

***

# Evolving traditions to suit a modern language

<div class="two-columns">
<div class="column">

-   Just strings → Proper data structures

    ```elvish
    var str = foo
    var list = [lorem [ipsum dolar]]
    var map = [&key=[&key=value]]
    ```

-   Text output → Value and text output (similar to PowerShell)

    ```elvish-transcript
    ~> put foo [lorem [ipsum dolar]]
    ▶ foo
    ▶ [lorem [ipsum dolar]]
    ```

-   Text pipeline → Value and text pipe

    ```elvish-transcript
    ~> put foobar "a\nb" | order
    ▶ "a\nb"
    ▶ foobar
    ```

</div>
<div class="column">

-   Exit code → Exception

    ```elvish-transcript
    ~> fail 'oh no'
    Exception: oh no
    [tty 36]:1:1-12: fail 'oh no'
    ~> false
    Exception: false exited with 1
    [tty 12]:1:1-5: false
    ```

</div>
</div>

***

# Functional programming

-   Traditional shells have FP in a sense...

    ```bash
    # Bash code
    trap 'echo exiting' EXIT
    PS1='$(git_prompt)'
    ```

-   Elvish has real lambdas, which look like `{ code }` or `{|arg| code}`

-   Once you have them, they pop up everywhere

    ```elvish
    peach {|h| ssh root@$h 'apt update'} [server-a server-b]
    time { sleep 1s }
    if (eq (uname) Darwin) { echo 'Hello macOS user!' }
    set edit:prompt = { print (whoami)@(tilde-abbr $pwd)'$ ' }
    ```

***

# Functional programming with immutable data

<div class="two-columns">
<div class="column">

-   Immutable data, not stateful objects

-   Strings, lists, maps are all immutable

    -   Make new values based on old values instead:

        ```elvish-transcript
        ~> conj [a b] c
        ▶ [a b c]
        ~> assoc [a b] 0 x
        ▶ [x b]
        ~> assoc [&k1=v1 &k2=v2] k1 new-v1
        ▶ [&k1=new-v1 &k2=v2]
        ```

    -   Lists and maps are based on
        [hash array mapped tries](https://en.wikipedia.org/wiki/Hash_array_mapped_trie)

</div>
<div class="column">

-   Variables (and variables alone) are mutable

    ```elvish
    var v = old
    set v = new
    ```

-   "Mutating" lists and maps

    ```elvish-transcript
    ~> var li = [a b]
    ~> var li2 = $li
    ~> set li2[0] = x
    ~> put $li $li2
    ▶ [a b]
    ▶ [x b]
    ```

</div>
</div>

***

# Functional programming with a concatenative twist

-   Pipelines model dataflows naturally

-   Nested functional calls:

    ```elvish
    var evens = [(keep-if {|x| == 0 (% $x 2)} [(range 100)])]
    ```

-   Concatenative:

    ```elvish
    range 100 | keep-if {|x| == 0 (% $x 2)} | var evens = [(all)]
    ```

-   We also get concurrency

***

# Namespaces and modules

-   Builtin modules:

    ```elvish
    use str
    str:has-prefix foobar f
    ```

-   Environment variables:

    ```elvish
    echo $E:LSCOLORS
    set E:http_proxy = ...
    ```

-   User-defined modules

    ```elvish
    # ~/.config/elvish/lib/a.elv
    var foo = bar
    # in shell session
    use a
    echo $a:foo
    ```

-   Local names are checked statically

    ```elvish
    set foo = bar # compilation error if foo is not declared
    ```

***

# Typing

-   Dynamic typing

    -   Scale down

    -   Interface with external data

-   Strong typing

    -   With some pockets of weak typing:

        ```elvish-transcript
        ~> + 1 2
        ▶ (num 3)
        ~> + (num 1) (num 2)
        ▶ (num 3)
        ```

    -   Potentially problematic:

        ```elvish
        var n = 0
        if (...) { set n = (+ $n 1) }
        ```

-   Static typing?

    -   Gradual typing and structural typing are essential

    -   TypeScript has got a lot right (but is way too big)

***

# Various bits I like

-   Lisp-1 vs Lisp-2:

    ```elvish
    echo a # resolves to $echo~
    var foo~ = { echo foo }
    foo
    ```

-   Arbitrary-precision numbers (R6RS numerical tower)

    ```elvish-transcript
    ~> * (range 1 40)
    ▶ (num 20397882081197443358640281739902897356800000000)
    ~> + 1/10 1/5
    ▶ (num 3/10)
    ```

-   Concatenation instead of interpolation:

    ```elvish
    echo 'Hello, '(whoami)
    ```

***

# What's next

-   TUI framework

-   Improve CI/CD experience with Elvish

-   Language features to figure out

    -   Modelling behavior of types

    -   Making exceptions work

    -   Dependency management

    -   Type system

-   The logo

***

# Cultural shifts

-   Elvish was started in 2013

-   Cultural changes in shell user community

    -   Skepticism for new shells has become less common

    -   "I'd rather just use Python/Ruby/...", because

        -   "I like my shell to be dumb"

        -   "I already know bash/zsh"

        -   The former has become less common over the years

-   Changes in programming culture

    -   C++/Java-style OOP is no longer canon

        -   Mutable objects are giving way to immutable data

        -   Inheritance has given way to composition and interfaces

    -   Exceptions have become less popular

    -   Gradual typing is now mainstream (implementations can still be
        hit-and-miss)

    -   Bar for tooling has risen (dependency management, editor support, ...)

***

# Review

-   What are unusual about Elvish all come from predecessors

    -   Traditional shell: syntax; data flow with pipelines and output capture

    -   PowerShell: IO can carry not just text

    -   Clojure: immutable data structures

    -   The result is a functional scripting language with a unique style

-   The rest is "just" general language design

    -   With some unique constraints

    -   And an ever-shifting goalpost

***

# Q&A
