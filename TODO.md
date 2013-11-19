# Next steps

* edit:
    - Highlighting beyond lexer and maybe even parser (highlighting of valid
      and invalid command names a la fish, and valid and invalid variable
      expansions)
    - Multiline editing
* eval:
    - Implement lexical scoping (variable capture of closures)
    - Determine namespacing model (importable modules; relationship of
      functions vs. external commands, variables vs. environmental variables)
    - Implement control structures; settle down its syntax - special syntax or
      closure syntax?
    - Determine failing model (boundary of failure; exception handling a la
      Python/Lua/...  and/or error value handling a la golang)
    - Support evaluating script
* All the TODO and XXX's in source :)
