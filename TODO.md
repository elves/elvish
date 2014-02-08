# Next steps

* Editor (edit)
    - Multiline editing
    - History
* Syntax (parse, eval)
    - Globbing?
    - Adopt value/address duality of `$a`, so that `set a []` is written as
      `set $a []` - like Perl, but unlike old shell or Tcl
    - Introduce the distinction of special forms vs. functions
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
* Data semantics, data passing (eval)
    - Determine mutability (mutable tables like in conventional imperative
      languages, or immutable data structure a la clojure?)
    - Allow functions to declare and use channel I/O
* Unix process stuff
    - Signal handling
    - Simple job control: support suspending and resuming jobs; no need for
      too complex job control, the role is which is largely subsumed by
      screen/tmux today
* All the TODO, XXX and BUG's in source :)
