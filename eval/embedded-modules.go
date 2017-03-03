package eval

var embeddedModules = map[string]string{
	"elves:acme": `fn acme {
    echo 'So this'
    put works.
}
`,
}
