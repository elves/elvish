# Next steps

* edit:
    - Send queries to and decode responses from terminal (deciding position of
      cursor when intializing editor, etc.)
    - Highlighting beyond lexer and maybe even parser (highlighting of valid
      and invalid command names, a la fish)
* eval:
    - Implement internal variable and lexical scoping - maybe just steal
      Python's local-closure-global scoping rule
    - Determine namespacing model (importable modules; relationship of
      functions vs. external commands, variables vs. environmental variables)
    - Implement control structures; settle down its syntax - special syntax or
      closure syntax?
    - Determine failing model (boundary of failure; exception handling a la
      Python/Lua/...  and/or error value handling a la golang)
    - Support evaluating script
* All the TODO and XXX's in source :)
