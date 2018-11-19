package main

import (
	"bytes"
	"testing"
)

var extractTests = []struct {
	name string
	src  string
	doc  string
}{
	{"Empty source", "", ""},
	{"Source without elvdoc", "package x\n// not elvdoc", ""},

	{"Source with doc-fn",
		`package x

//elvdoc:fn cd
//
// Changes directory.
`,
		`# Functions

## cd

Changes directory.
`},

	{"Source with multiple doc-fn",
		`package x

//elvdoc:fn b
// B.

//elvdoc:fn a
// A.

//elvdoc:fn c
// C.
`,
		`# Functions

## a
A.

## b
B.

## c
C.
`,
	},

	{"Source with both doc-fn and var-fn",
		`package x

//elvdoc:fn a
// A.

//elvdoc:var $b
// B.
`,
		`# Variables

## $b
B.


# Functions

## a
A.
`,
	},

	{"Source without trailing newline",
		`package x

//elvdoc:fn a
// A.`, `# Functions

## a
A.
`,
	},
}

func TestExtract(t *testing.T) {
	for _, test := range extractTests {
		t.Run(test.name, func(t *testing.T) {
			r := bytes.NewBufferString(test.src)
			w := new(bytes.Buffer)
			extract(r, w)
			if w.String() != test.doc {
				t.Errorf("\n<<<<< Got\n%s\n=====\n%s\n>>>>> Want",
					w.String(), test.doc)
			}
		})
	}
}
