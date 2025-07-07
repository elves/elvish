/*
Package etk implements a TUI framework.
It has two prominent features:

  - It provides what we call an "immediate mode" API,
    in line with modern UI frameworks,
    such as [React], [SwiftUI] and [Jetpack Compose].
    (See notes below for more details.)

    This style of API lets you define a function that describes the UI,
    and the framework will call it whenever there is an update.
    The lifecycle of an app is managed by the framework.

    The process of calling the component function is known by various names,
    including "rendering", "recomposition" and "view update".
    In Etk we can it "rendering".

  - It manages all the persistent state of the UI in a tree structure,
    implemented as an Elvish map,
    with the state of each subcomponent as a nested map.

    Each component can freely manipulate its subtree -
    that is, the nested map storing its own states and those of its descendants.

    The state tree is the only way to pass information in Etk;
    if a component has any customizable property,
    it is modelled as a state variable for its parent to modify.
    (For those familiar with React,
    states in Etk correspond to both "states" and "props" in React.)

An Etk component is implemented by a [Comp]:
a function taking a [Context],
which provides access to the state subtree
(mediated by [StateVar]'s),
and returns a [View] and a [React]:

  - The View describes the UI's appearance.

  - The React describes the UI's behavior -
    it's a function that will be called to react when an event happens.

[Run] runs a Comp;
it implements an application main loop,
taking care of the lifecycle of the application.

For more technical details,
read the respective documentation of the symbols above.

# Implementation structure

Most of the implementation is split as follows:

  - etk.go defines all the core types except View
    ([Comp], [Context], [StateVar] and [React])
    and utilities to manipulate them.

  - view.go defines the [View] interface and its various implementations;
    it's relatively independent from the other core types.

  - run.go implements [Run].

# Testing Etk components

The [src.elv.sh/pkg/etk/etktest] package supports testing Etk components.

# Notes on "immediate mode" API

Most traditional GUI APIs are so-called "retained mode".
You start by creating a bunch of components;
the framework "retains" them for you,
and lets you modify them later.
The browser's DOM and many other traditional GUI frameworks follow this model.
You can think of the application state as a canvas,
and you're programming pencils and erasers that add or remove lines in it.

In retained mode,
complex business logic can often result in complex ways of modifying components,
making it hard to ensure that the properties of different components -
and indeed, which components are shown at the same time -
are always kept consistent.
As a result, "immediate mode" API has become more popular recently.
They are modelled after the "immediate mode" of computer graphics,
where for each frame you render the entire screen from scratch.
Think of a CRT screen, where the screen has to be constantly refreshed.

Similarly, with an immediate-mode GUI framework,
instead of creating components and modify them later,
you declare a function that render all the components from scratch every time.
This makes it much easier to reason about the relationship between components.

There are some pseudo-immediate mode GUI frameworks built from scratch,
like [Dear ImGui] and [Gio UI].

On the other hand,
[React], [SwiftUI] and [Jetpack Compose] provide immediate mode APIs above an
underlying [retained mode] API.

As good as it sounds,
immediate-mode frameworks still have to store the state of the application
somewhere,
and it is in this aspect they differ the most.
Etk uses a "managed but open" Elvish map to make applications created with Etk
maximally customizable using Elvish code.

[retained mode]: https://en.wikipedia.org/wiki/Retained_mode
[Dear ImGui]: https://github.com/ocornut/imgui
[Gio UI]: https://gioui.org
[React]: https://react.dev
[SwiftUI]: https://developer.apple.com/xcode/swiftui/
[Jetpack Compose]: https://developer.android.com/compose

[immediate mode]: https://en.wikipedia.org/wiki/Immediate_mode_(computer_graphics)
*/
package etk
