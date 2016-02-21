package eval

var builtinModules = map[string]string{
	"acme": `fn acme {
    echo 'So this'
    put works.
}
`,
}
