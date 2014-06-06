# Next steps

* Editor (edit)
    - Record mode for saving part of shell session as scripts
    - Interactive debugger (steal some ideas from LightTable)
    - Many...
* Syntax (parse, eval)
    - Sigil as shorthand for function invocation
        + `vim =yaourt` is equivalent to `vim (= yaourt)`
        + Implement globbing as a sigil to avoid cluttering the lexical
          structure (which sigil to use?)
        + Sigils are lexically scoped (they are just functions after all)
        + A closed set of characters eligible as sigils
* Code structure (eval)
    - TYPE CHECKER
    - Implement control structures - as special syntax or high-order function?
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
    - Intermediate state:
        ```
        var $a $b = (echo a | tee /tmp/a)
        ```
      The command fails, but /tmp/a is written
* Simple job control
    - Support suspending and resuming jobs
    - No need for too complex job control, the role is which is largely
      subsumed by screen/tmux today
* All the TODO, XXX and BUG's in source :)
