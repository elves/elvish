package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/util"
)

var extractTests = []struct {
	name    string
	src     string
	ns      string
	wantDoc string
}{
	{name: "Empty source", src: "", wantDoc: ""},
	{name: "Source without elvdoc", src: "package x\n// not elvdoc", wantDoc: ""},

	{
		name: "Source with doc-fn",
		src: `package x

//elvdoc:fn cd
//
// Changes directory.
`,
		wantDoc: `# Functions

## cd

Changes directory.
`,
	},

	{
		name: "Source with multiple doc-fn",
		src: `package x

//elvdoc:fn b
// B.

//elvdoc:fn a
// A.

//elvdoc:fn c
// C.
`,
		wantDoc: `# Functions

## a
A.

## b
B.

## c
C.
`,
	},

	{
		name: "Source with both doc-fn and var-fn",
		src: `package x

//elvdoc:fn a
// A.

//elvdoc:var b
// B.
`,
		wantDoc: `# Variables

## $b
B.


# Functions

## a
A.
`,
	},

	{
		name: "Source without trailing newline",
		src: `package x

//elvdoc:fn a
// A.`,
		wantDoc: `# Functions

## a
A.
`,
	},
	{
		name: "Source with both doc-fn and var-fn",
		src: `package x

//elvdoc:fn a
// A.

//elvdoc:var b
// B.
`,
		ns: "ns:",
		wantDoc: `# Variables

## $ns:b
B.


# Functions

## ns:a
A.
`,
	}}

func TestExtract(t *testing.T) {
	for _, test := range extractTests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.src)
			w := new(strings.Builder)
			extract(r, test.ns, w)
			compare(t, w.String(), test.wantDoc)
		})
	}
}

func TestRun_MultipleFiles(t *testing.T) {
	teardown := setup()
	defer teardown()

	w := new(strings.Builder)
	run([]string{"a.go", "b.go"}, emptyReader, w)
	compare(t, w.String(), `# Variables

## $v2

Variable 2 from b.


# Functions

## f1

Function 1 from b.

## f2

Function 2 from a.
`)
}

func TestRun_Directory(t *testing.T) {
	teardown := setup()
	defer teardown()

	w := new(strings.Builder)
	run([]string{"-dir", "."}, emptyReader, w)
	compare(t, w.String(), `# Variables

## $v1

Variable 1 from c.

## $v2

Variable 2 from b.


# Functions

## f1

Function 1 from b.

## f2

Function 2 from a.
`)
}

func TestRun_Filter(t *testing.T) {
	teardown := setup()
	defer teardown()

	in := strings.NewReader(`Some text.

$elvdoc a.go

Some more text.`)
	out := new(strings.Builder)
	run([]string{"-filter"}, in, out)
	compare(t, out.String(), `Some text.

# Functions

## f2

Function 2 from a.

Some more text.
`)
}

func compare(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("\n<<<<< Got\n%s\n=====\n%s\n>>>>> Want", got, want)
	}
}

// Set up a temporary directory with several .go files and directories
// containing .go files. Returns a teardown function. Useful for testing the run
// function.
func setup() func() {
	_, teardown := util.InTestDir()
	writeFile("a.go", `package x
//elvdoc:fn f2
//
// Function 2 from a.
`)
	writeFile("b.go", `package x
//elvdoc:fn f1
//
// Function 1 from b.

//elvdoc:var v2
//
// Variable 2 from b.
`)
	writeFile("c.go", `package x
//elvdoc:var v1
//
// Variable 1 from c.
`)
	writeFile("notgo.gox", `package x
//elvdoc:var wontappear
//
// This won't appear because it is not in a .go file.
`)
	// Subdirectories are ignored with -dir.
	writeFile("subpkg/a.go", `package subpkg
//elvdoc:fn subpkg:f
//
// Function f from subpkg/a.

//elvdoc:var subpkg:v
//
// Variable v from subpkg/a.
`)
	return teardown
}

func writeFile(name, data string) {
	dir := path.Dir(name)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(name, []byte(data), 0600)
}
