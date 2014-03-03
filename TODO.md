# Next steps

* Editor (edit)
    - Multiline editing
* Syntax (parse, eval)
    - Sigil as shorthand for function invocation
        + `vim =yaourt` is equivalent to `vim (= yaourt)`
        + Implement globbing as a sigil to avoid cluttering the lexical
          structure (which sigil to use?)
        + Sigils are lexically scoped (they are just functions after all)
        + A closed set of characters eligible as sigils
* Code structure (eval)
    - STATIC ANALYSER AND TYPE CHECKER
    - Implement variable capturing of closures, so that the following sets
      `$v`:
        ```
        var $v
        { set $v 2 }
        ```
    - Implement control structures - as special syntax or high-order function?
    - Support evaluating script
    - Determine namespacing mechanism (importable modules; relationship of
      functions vs. external commands, variables vs. environmental variables)
* Value semantics and data passing (eval)
    - First class IO values
    - Immutable by default, a la Clojure
        + Believed to faciliate concurrency
        + Just makes more sense
        + But namespace and $env need to be mutable - $env as a namespace?
    - Allow functions to declare and use channel I/O
* Failing/exception/error (eval)
    - Determine failing mechanism
        + Exception handling a la Python/Lua/...?
        + Error parallel to value a la Icon?
        + Error as return value a la golang?
        + Error as special $status variable a la old Shell?
    - Failing behavior of builtins; intermediate state:
        ```
        var (echo a | tee /tmp/a) b = 1
        ```
      Should arity mismatch be detected early and avoid the side-effect of
      writing `/tmp/a`? Should $a and $b be defined?
      (`var` is going to be a special form, needs another subtle example)
* Unix process stuff
    - Signal handling
    - Simple job control: support suspending and resuming jobs; no need for
      too complex job control, the role is which is largely subsumed by
      screen/tmux today
* All the TODO, XXX and BUG's in source :)
