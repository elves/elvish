# Designing Etk, a bilingual TUI framework with unusual constraints and trade-offs

Qi Xiao (xiaq)

2024-09-19 @ Recurse Center

***

# Background

-   Elvish is this shell that I develop with a lot of TUI features

-   The TUI framework is one of the main projects I did at RC

-   Not quite complete, but design questions are mostly settled

***

# Elvish's old TUI framework

-   Built around the `Widget` interface:

    ```go
    type Widget interface {
        Render(width, height int) *term.Buffer
        Handle(event term.Event) bool
    }
    ```

-   Dive into examples: `pkg/cli/tk/codearea.go`, `pkg/cli/tk/combobox.go`

    -   Initialization

    -   Composition

    -   Additional behavior (including state mutation)

***

# Limitations and motivations for new framework

-   Composition and state management are repetitive

-   Mutating a set of related UI states is error-prone

-   Each widget needs a bespoke Elvish API

-   No way to program the TUI using Elvish code

-   New framework!

    -   Opionated in composition and state management

    -   Immediate mode rather than retained mode

    -   Expose an Elvish API automatically

    -   Programmable from Elvish

***

# Idea 1: Immediate mode

-   Retained mode: build an initial UI, mutate it on demand:

    ```
    <input type="checkbox" id="foo"></input>
    <input type="checkbox" id="bar"></input>
    on('some-event', () => {
      $('#foo').checked = fooExpr;
      $('#bar').checked = barExpr;
    });
    ```

-   Immediate mode: re-generate the UI every time:

    ```
    function myComp({}) {
      return <>
        <input type="checkbox" checked={fooExpr}></input>
        <input type="checkbox" checked={barExpr}></input>
      </>
    }
    ```

-   Originates from graphics API

-   Most modern UI frameworks have adopted this style

***

# Idea 2: State map

-   Immediate mode doesn't solve the problem of state management

-   Different approaches are possible

    -   Just put it in global variables

    -   [The Elm architecture](https://guide.elm-lang.org/architecture/): Typed
        models and messages

-   I want automatic Elvish API for components

    -   Use a nested map

***

# Putting it together: Etk

-   A component is

    -   A function in Go or Elvish

    -   That takes a map containing the current state

    -   And outputs a UI and an event handler

***

# Build some components

***

# Basic: Counter

-   Most basic app from [7GUIs](https://eugenkiss.github.io/7guis/tasks/)

-   Go: `pkg/etk/examples/counter.go`

-   Elvish: `pkg/etk/examples/counter.elv`

***

# Subcomponents: Temperature conversion

-   Go: `pkg/etk/examples/temperature.go`

-   Initializing, reading and writing the state of subcomponents

***

# Subcomponents are states

-   The subcomponent function is a state

-   Subcomponent's state is state

-   State-showing wrapper

***

# Reviewing the design

***

# Overall style

-   Immediate mode

-   Function + managed state

-   A funky isomorphism of the Elm architecture?

***

# Open state tree

-   State is inspectable and tinkerable

-   Not a sealed "product"

-   No encapsulation

    -   A component can mutate the state of any of its descendant

    -   Access to the root allows you to inspect and mutate any point of the
        state tree

    -   Very much working as intended

-   Unsafe

    -   State bindings are based on string names and not type-safe

    -   Not ideal, but necessary for integrating with a dynamic language like
        Elvish

***

# More things we can do with the state API

-   All state access is via an API

-   We can get concurrency safety mostly for free

-   Undo/redo mostly for free

-   (To be implemented)

***

# La fin

<!--
# Implementation 

***

# TUI primitives

-   Terminal is mostly text

-   In-band signals: escape codes

    -   Combination and function keys

    -   Text style

    -   Cursor addressing

-   Out-of-band control: signals and `ioctl`

-   Unix's terminal API dates back to the 1960s (TTY = **t**ele**ty**pewritter)

    -   Various unsuccessful reform attempts throughout history

***

# What I can't explain

-   A component complects two things: view and react (it's in the signature)

-   Some frameworks keep them separate (Elm, Bubble Tea, "class-based" React)

-   But some complect them ("function-based" React)

-   I can't quite explain why I like the latter, other than "it keeps a
    component a single thing"

***
-->
