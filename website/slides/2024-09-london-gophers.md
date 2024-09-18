# How to test your programming language by inventing a DSL and a VS Code plugin

Qi Xiao (xiaq)

2024-09-18 @ London Gophers

***

# Intro

-   About myself

-   About the programming language we're testing

    -   Elvish - a programming language, but also a modern shell

-   Like bash / zsh / ..., but more modern

    -   More powerful interactive features

    -   Full-fledged programming language

    -   Other modern shells: [Nushell](https://www.nushell.sh),
        [Oils](https://www.oilshell.org), [Murex](https://murex.rocks)

***

# Full-fledged programming language

-   Some think advanced programming features and shell scripting are
    incompatible

-   But real programming features are great for shell scripting!

    ```elvish
    # [foo bar] - list
    # [&key=value] - map
    var hosts = [[&name=a &cmd='apt update']
                 [&name=b &cmd='pacman -Syu']]
    # peach = "parallel each"
    # {|h| ...} - lambda
    peach {|h| ssh root@$h[name] $h[cmd] } $hosts
    ```

-   Elvish has all the familiar shell features too

    ```elvish
    vim main.go
    cat *.go | wc -l
    # Elvish also supports recursive wildcards
    cat **.go | wc -l
    ```

***

# Interactive features

-   Great out-of-the-box experience (demo)

    -   Syntax highlighting

    -   Completion with <kbd>Tab</kbd>

    -   Directory history with <kbd>Ctrl-L</kbd>

    -   Command history with <kbd>Ctrl-R</kbd>

    -   Filesystem navigator with <kbd>Ctrl-N</kbd>

-   Programmable

    ```elvish
    set edit:prompt = { print (whoami)@(tilde-abbr $pwd)'$ ' }
    ```

***

# Testing the Elvish interpreter

***

# Test strategy

-   Testing is important

    -   Gives us confidence about the correctness of the code

    -   Especially when changing the code

-   Most important thing about your test strategy

    -   Make it *really* easy to create and maintain tests

    -   Easy-to-write tests ⇒ more tests ⇒ higher test coverage

    -   Elvish has 92% test coverage

-   Interpreters have a super simple API!

    -   Input: code

    -   Output: text, values

        ```elvish-transcript
        ~> echo hello world
        hello world
        ~> put [hello world] [foo bar]
        ▶ [hello world]
        ▶ [foo bar]
        ```

***

# Iteration 1: table-driven tests

```go
// Simplified interpreter API
func Interpret(code string) ([]any, string)

var tests = []struct{
    code string
    wantValues []any
    wantText   string
}{
    {code: "echo foo", wantText: "foo\n"},
}

func TestInterpreter(t *testing.T) {
    for _, test := range tests {
        gotValues, gotText := Interpret(test.code)
        // Compare with test.wantValues and test.wantText
    }
}
```

***

# Adding a test case with table-driven tests

-   Steps:

    1.  Implement new functionality

    2.  Test manually in terminal:

        ```elvish-transcript
        ~> str:join , [a b]
        ▶ 'a,b'
        ```

    3.  Convert the interaction into a test case:

        ```go
        {code: "str:join , [a b]", wantValues: []any{"a,b"}}
        ```

-   Step 3 can get repetitive

    -   Computers are good at repetitive tasks 🤔

***

# Iteration 2: transcript tests

-   Record terminal *transcripts* in `tests.elvts`:

    ```elvish-transcript
    ~> str:join , [a b]
    ▶ 'a,b'
    ```

-   Generate the table from the terminal transcript:

    ```go
    //go:embed tests.elvts
    const transcripts string

    func TestInterpreter(t *testing.T) {
        tests := parseTranscripts(transcripts)
        for _, test := range tests { /* ... */ }
    }
    ```

-   Embrace text format

    -   We lose strict structure, but it doesn't matter in practice

***

# Adding a test case with transcript tests

-   Steps:

    1.  Implement new functionality

    2.  Test manually in terminal:

        ```elvish-transcript
        ~> str:join , [a b]
        ▶ 'a,b'
        ```

    3.  Copy the terminal transcript into `tests.elvts`

-   Copying is still work

    -   What if we don't even need to copy? 🤔

***

# Iteration 2.1: an editor extension for transcript tests

-   Editor extension for `.elvts` files

    -   Run code under cursor

    -   Insert output below cursor

-   Steps (demo):

    1.  Implement new functionality

    2.  Test manually in `tests.elvts` within the editor:

        ```elvish-transcript
        ~> use str
        ~> str:join , [a b]
        ▶ 'a,b'
        ```

-   We have eliminated test writing as a separate step during development!

***

# Tangent: a weird dependency injection trick

<div class="two-columns">
<div class="column">

You're probably familiar with dependency injection tricks like this:

```go
// in foo.go
package foo
var stdout = os.Stdout
func Hello() {
    fmt.Fprintln(stdout, "Hello!")
}

// in foo_test.go
package foo
func TestHello(t *testing.T) {
    stdout = ...
    ...
}
```

</div>
<div class="column">

What if the test is an external test? You can export `stdout`, but that makes it
part of the API. Instead:

```go
// foo.go is unchanged

// in testexport_test.go
package foo // an internal test file
var Stdout = &stdout

// in foo_test.go
package foo_test // an external test file
func TestHello(t *testing.T) {
    *foo.Stdout = ...
    ...
}
```

</div>
</div>

***

# Testing the terminal app

***

# Widget abstraction

-   Like GUI apps, Elvish's terminal app is made up of *widgets*

    ```go
    type Widget interface {
        Handle(event Event)
        Render(width, height int) *Buffer
    }
    ```

-   `Buffer`: stores *rich text* and the cursor position

-   `Event`: keyboard events (among others)

-   Example: `CodeArea`

    -   Stores text content and cursor position

    -   `Render`: writes a `Buffer` with current content and cursor

    -   `Handle`:

        -   <kbd>a</kbd> → insert `a`

        -   <kbd>Backspace</kbd> → delete character left of cursor

        -   <kbd>Left</kbd> → move cursor left

***

# Widget API is also simple(-ish)

-   Input: `Event`

-   Output: `Buffer`

-   But:

    -   Multiple inputs and outputs, often interleaved.

        A typical test:

        1.  Press <kbd>x</kbd>, press <kbd>y</kbd>, render and check

        2.  Press <kbd>Left</kbd>, render and check

        3.  Press <kbd>Backspace</kbd>, render and check

    -   Tests end up verbose and not easy to write 😞

***

# Leveraging Elvish and transcript tests!

-   Create Elvish bindings for the widget

-   Now just use Elvish transcript tests

    ```elvish-transcript
    ~> send [x y]; render
    xy
    ~> send [Left]; render
    xy
    ~> send [Backspace]; render
    y
    ```

-   Look a lot like screenshots tests!

    -   With "screenshots" embedded directly in test files

***

# Encoding text style and cursor position

Actual `render` output is slightly more sophisticated:

```elvish-transcript
~> send [e c o]; render
┌────────────────────────────────────────┐
│eco                                     │
│RRR ̅̂                                    │
└────────────────────────────────────────┘
~> send [Left]; render
┌────────────────────────────────────────┐
│eco                                     │
│RRR̅̂                                     │
└────────────────────────────────────────┘
~> send [h]; render
┌────────────────────────────────────────┐
│echo                                    │
│GGGG̅̂                                    │
└────────────────────────────────────────┘
```

***

# Can this be even easier?

-   We still need to manually transcribe our testing session

-   Next step: record actual TUI sessions?

***

# Conclusions

***

# Elvish's testing strategy

-   Make testing easy

-   Embrace text, embrace the editor

-   Prior art: [Mercurial's tests](https://wiki.mercurial-scm.org/WritingTests)

***

# Plugging Elvish once more

-   Use Elvish: <https://elv.sh/>

    -   Get Elvish: <https://elv.sh/get/> (one-liner installation script thanks
        to Go)

    -   Adopting a shell is not an "all or nothing" matter

    -   Try Elvish in the browser: <https://try.elv.sh>

-   Hack on Elvish: <https://github.com/elves/elvish>

    -   Developer docs: <https://github.com/elves/elvish/tree/master/docs>

***

# Q&A
