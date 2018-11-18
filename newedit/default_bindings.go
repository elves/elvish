package newedit

// Elvish code for default bindings, assuming the editor ns as the global ns.
const defaultBindingsElv = `
insert:binding = (binding-map [
  &Ctrl-D=  $commit-eof~
  &Default= $insert:default-handler~
])
`

// vi: set et:
