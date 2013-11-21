# Next steps

* Editor (edit)
    - Tab completion
    - Multiline editing
* Syntactic constructs
    - Globbing?
* General/misc runtime feature (eval)
    - Implement variable capture of closures
    - Implement control structures; settle down its syntax - special syntax or
      closure syntax?
    - Support evaluating script
    - Determine namespacing mechanism (importable modules; relationship of
      functions vs. external commands, variables vs. environmental variables)
* Failing/exception/error
    - Determine failing mechanism
        + Exception handling a la Python/Lua/...?
        + Error as return value a la golang?
        + Error as special $status variable?
    - Failing behavior of builtins; intermediate state:
        ```
        var ${echo a | tee /tmp/a} b = 1
        ```
      Should arity mismatch be detected early and avoid the side-effect of
      writing `/tmp/a`? Should $a and $b be defined?
* Data semantics, data passing (eval)
    - Determine mutability (mutable tables like in conventional imperative
      languages, or immutable data structure a la clojure?)
    - Adopt the reference/value duality of $a, so that `set a []` is
      written as `set $a []` (like Perl)? Would faciliates variable capture of
      closures
    - Implement output capture: ${}
    - Allow functions to declare and use channel I/O
* Unix process stuff
    - Signal handling
    - Simple job control: support suspending and resuming jobs; no need for
      too complex job control, the role is which is largely subsumed by
      screen/tmux today
* All the TODO and XXX's in source :)
