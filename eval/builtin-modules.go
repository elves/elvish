package eval

var builtinModules = map[string]string{
	"elves:acme": `fn acme {
    echo 'So this'
    put works.
}
`,
}
