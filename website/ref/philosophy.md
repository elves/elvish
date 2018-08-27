<!-- toc -->

The development of Elvish is driven by a set of ideas, a **design philosophy**.

# The language

*   Elvish should be a real, expressive programming language.

    Shells are often considered domain-specific languages (DSL), but Elvish
    does not restrict itself to this notion. It embraces such concepts as
    namespaces, first-class functions and exceptions. Whatever you may find in
    a modern general-purpose programming language is likely to be found in
    Elvish.

    Elvish is not alone in this respect. There are multiple ongoing efforts;
    [this page](https://github.com/oilshell/oil/wiki/ExternalResources) on
    the wiki of oilshell (which is one of the efforts) is a good reference.

*   Elvish should try to preserve and extend traditional shell programming
    techniques, as long as they don't conflict with the previous tenet. Some
    examples are:

    *   Barewords are simply strings.

    *   Prefix notation dominates, like Lisp. For example, arithmetics is done
        like `+ 10 (/ 105 5)`.

    *   Pipeline is the main tool for function composition. To make pipelines
        suitable for complex data manipulation, Elvish extends them to be able
        to carry structured data (as opposed to just bytes).

    *   Output capture is the auxiliary tool for function composition. Elvish
        functions may write structured data directly to the output, and
        capturing the output yields the same structured data.

# The user interface

*   The user interface should be usable without any customizations. It should
    be simple and consistent by default:

    *   Prefer to extend well-known functionalities in other shell to inventing
        brand new ones. For instance, in Elvish Ctrl-R summons the "history
        listing" for searching history, akin to how Ctrl-R works in bash, but
        more powerful.

    *   When a useful feature has no prior art in other shells, borrow from
        other programs. For instance, the [navigation
        mode](/learn/cookbook.html#navigation-mode), summoned by Ctrl-N,
        mimics [Ranger](http://ranger.nongnu.org); while the "location mode"
        used for quickly changing location, mimics location bars in GUI
        browsers (and is summoned by the same key combination Ctrl-L).

*   Customizability should be achieved via progammability, not an enormous
    inventory of options that interact with each other in obscure ways.
