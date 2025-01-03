# Implementing and testing a shell and programming language in Go

Qi Xiao (xiaq)

2024-10-02 @ Golang Oxford

<!-- This talk is similar in scope to 2024-08-gophercon-uk.md, but using the
revised content from 2024-09-london-gophers.md as the starting point,
re-integrating and revising parts of the implementation section. Notably, the
parsing and compilation parts are dropped entirely, and we only focus on the
shell-specific and Go-specific areas. -->

***

# Intro

-   About myself

-   The programming language and shell this talk is about

    -   Elvish <https://elv.sh>

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

# Implementing the Elvish interpreter

***

# Parsing and "compiling"

-   Source code

    ```elvish
    echo $pid | wc
    ```

<div class="two-columns">
<div class="column">

-   Syntax tree:

    ![AST](./2024-08-rc-implementation/syntax-tree.svg)

    <!--
    digraph AST {
        node [shape = rectangle];
        edge [arrowhead = none];
        "Form echo" [label = "Form"];
        "Form wc" [label = "Form"];
        "Expr echo" [label = "Expr\nType=Bareword\nValue=\"echo\""];
        "Expr $pid" [label = "Expr\nType=Variable\nValue=\"pid\""];
        "Expr wc" [label = "Expr\nType=Bareword\nValue=\"wc\""];

        "Pipeline" -> "Form echo" [label = "Form"];
        "Form echo" -> "Expr echo" [label = "Head"];
        "Form echo" -> "Expr $pid" [label = "Arg"];
        "Pipeline" -> "Form wc" [label = "Form"];
        "Form wc" -> "Expr wc" [label = "Head"];
    }
    -->

</div>
<div class="column">

-   Op tree:

    ![Op tree](./2024-08-rc-implementation/op-tree.svg)

    <!--
    digraph optree {
        node [shape = rectangle];
        edge [arrowhead = none];
        "Pipeline" [label = "pipelineOp"];
        "Form echo" [label = "formOp"];
        "Form wc" [label = "formOp"];
        "Expr echo" [label = <<b>variableOp</b><br/>Scope=<b>Builtin</b><br/>Name=<b>"echo~"</b>>];
        "Expr $pid" [label = <variableOp<br/>Scope=<b>Builtin</b><br/>Name="pid">];
        "Expr wc" [label = <<b>literalOp</b><br/>Value=<b>ExternalCmd{"wc"}</b>>];

        "Pipeline" -> "Form echo" [label = "Form"];
        "Form echo" -> "Expr echo" [label = "Head"];
        "Form echo" -> "Expr $pid" [label = "Arg"];
        "Pipeline" -> "Form wc" [label = "Form"];
        "Form wc" -> "Expr wc" [label = "Head"];
    }
    -->

</div>
</div>

***

# Execution

-   The `exec` method is where the real action happens

    ```go
    type pipelineOp struct { formOps []formOp }
    func (op *pipelineOp) exec() { /* ... */ }

    type formOp struct { /* ... */ }
    func (op *formOp) exec() { /* ... */ }
    ```

-   ```elvish
    echo $pid | wc
    ```

    How do we connect the output of `echo` to the input of `wc`?

-   You exist in the context of all in which you live

    ```go
    type Context struct {
        stdinFile *os.File; stdinChan <-chan any
        stdoutFile *os.File; stdoutChan chan<- any
    }

    func (op *pipelineOp) exec(*Context) { /* ... */ }
    func (op *formOp) exec(*Context) { /* ... */ }
    ```

***

# Executing a pipeline

```go
type pipelineOp struct { forms []formOp }

func (op *pipelineOp) exec(ctx *Context) {
    form1, form2 := forms[0], forms[1] // Assume 2 forms
    r, w, _ := os.Pipe()               // Byte pipeline
    ch := make(chan any, 1024)         // Channel pipeline
    ctx1 := ctx.cloneWithStdout(w, ch) // Context for form 1
    ctx2 := ctx.cloneWithStdin(r, ch)  // Context for form 2
    var wg sync.WaitGroup              // Now execute them in parallel!
    wg.Add(2)
    go func() { form1.exec(ctx1); wg.Done() }()
    go func() { form2.exec(ctx2); wg.Done() }()
    wg.Wait()
}
```

-   [Real code](https://github.com/elves/elvish/blob/d8e2284e61665cb540fd30536c3007c4ee8ea48a/pkg/eval/compile_effect.go#L69)

***

# Go is great for writing a shell

-   Pipeline semantics

    -   Text pipelines: [`os.Pipe`](https://pkg.go.dev/os#Pipe)

    -   Value pipelines: channels

    -   Concurrent execution: Goroutines and
        [`sync.WaitGroup`](https://pkg.go.dev/sync)

-   Running external commands:
    [`os.StartProcess`](https://pkg.go.dev/os#StartProcess)

***

# Go is great for writing an interpreted language

-   Rich standard library

    -   Big numbers ([`big.Int`](https://pkg.go.dev/math/big#Int) and
        [`big.Rat`](https://pkg.go.dev/math/big#Rat)):

        ```elvish-transcript
        ~> * (range 1 41) # 40!
        ▶ (num 815915283247897734345611269596115894272000000000)
        ~> + 1/10 2/10
        ▶ (num 3/10)
        ```

    -   [`math`](https://pkg.go.dev/math),
        [`strings`](https://pkg.go.dev/strings) (`str:` in Elvish),
        [`regexp`](https://pkg.go.dev/regexp) (`re:` in Elvish):

        ```elvish-transcript
        ~> math:log10 100
        ▶ (num 2.0)
        ~> str:has-prefix foobar foo
        ▶ $true
        ~> re:match '^foo' foobar
        ▶ $true
        ```

-   Garbage collection comes for free!

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
