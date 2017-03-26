package eval

var embeddedModules = map[string]string{
	"embedded:acme": `fn acme {
    echo 'So this'
    put works.
}
`,
}
