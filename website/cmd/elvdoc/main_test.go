package main

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"src.elv.sh/pkg/testutil"
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
		name: "Source with elvdoc:fn",
		src: `package x

//elvdoc:fn cd
//
// Changes directory.
`,
		wantDoc: `# Functions

<a name='//apple_ref/cpp/Function/cd' class='dashAnchor'></a>

## cd {#cd}

Changes directory.
`,
	},

	{
		name: "symbol with punctuation and specified ID",
		src: `package x

//elvdoc:fn + {#add}
//
// Add.
`,
		wantDoc: `# Functions

<a name='//apple_ref/cpp/Function/%2B' class='dashAnchor'></a>

## + {#add}

Add.
`,
	},

	{
		name: "Source with unstable symbols",
		src: `package x

//elvdoc:fn -b
// -B.

//elvdoc:fn a
// A.

//elvdoc:fn b
// B.
`,
		ns: "ns:",
		wantDoc: `# Functions

<a name='//apple_ref/cpp/Function/ns%3Aa' class='dashAnchor'></a>

## ns:a {#ns:a}
A.

<a name='//apple_ref/cpp/Function/ns%3Ab' class='dashAnchor'></a>

## ns:b {#ns:b}
B.

<a name='//apple_ref/cpp/Function/ns%3A-b' class='dashAnchor'></a>

## ns:-b {#ns:-b}
-B.
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

<a name='//apple_ref/cpp/Function/a' class='dashAnchor'></a>

## a {#a}
A.

<a name='//apple_ref/cpp/Function/b' class='dashAnchor'></a>

## b {#b}
B.

<a name='//apple_ref/cpp/Function/c' class='dashAnchor'></a>

## c {#c}
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

<a name='//apple_ref/cpp/Variable/%24b' class='dashAnchor'></a>

## $b {#b}
B.


# Functions

<a name='//apple_ref/cpp/Function/a' class='dashAnchor'></a>

## a {#a}
A.
`,
	},

	{
		name: "Elvish source",
		src: `
#elvdoc:fn a
# A.

#elvdoc:var b
# B.
`,
		wantDoc: `# Variables

<a name='//apple_ref/cpp/Variable/%24b' class='dashAnchor'></a>

## $b {#b}
B.


# Functions

<a name='//apple_ref/cpp/Function/a' class='dashAnchor'></a>

## a {#a}
A.
`,
	},

	{
		name: "Source without trailing newline",
		src: `package x

//elvdoc:fn a
// A.`,
		wantDoc: `# Functions

<a name='//apple_ref/cpp/Function/a' class='dashAnchor'></a>

## a {#a}
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

<a name='//apple_ref/cpp/Variable/%24ns%3Ab' class='dashAnchor'></a>

## $ns:b {#ns:b}
B.


# Functions

<a name='//apple_ref/cpp/Function/ns%3Aa' class='dashAnchor'></a>

## ns:a {#ns:a}
A.
`,
	}}

var emptyReader = io.MultiReader()

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
	setupDir(t)

	w := new(strings.Builder)
	run([]string{"a.go", "b.go"}, emptyReader, w)
	compare(t, w.String(), `# Variables

<a name='//apple_ref/cpp/Variable/%24v2' class='dashAnchor'></a>

## $v2 {#v2}

Variable 2 from b.


# Functions

<a name='//apple_ref/cpp/Function/f1' class='dashAnchor'></a>

## f1 {#f1}

Function 1 from b.

<a name='//apple_ref/cpp/Function/f2' class='dashAnchor'></a>

## f2 {#f2}

Function 2 from a.

    Some indented code.
`)
}

func compare(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("\n<<<<< Got\n%s\n=====\n%s\n>>>>> Want", got, want)
	}
}

// Set up a temporary directory with several .go files and directories
// containing .go files.
func setupDir(c testutil.Cleanuper) {
	testutil.InTempDir(c)
	writeFile("a.go", `package x
//elvdoc:fn f2
//
// Function 2 from a.
//
//     Some indented code.
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
}

func writeFile(name, data string) {
	dir := path.Dir(name)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		panic(err)
	}
	os.WriteFile(name, []byte(data), 0600)
}
