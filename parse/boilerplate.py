#!/usr/bin/python2.7
"""
Generate helper functions for node types.

For every struct type T, it generates two functions:

* A IsT func that reports whether a node has type *T.

* A GetT func that takes Node and returns *T. It examines whether the Node is
  actually of type *T, and if it is, returns it; otherwise it returns nil.

For example, for the following type:

type X struct {
    node
    F *Y
    G *[]Z
}

The following boilerplate is generated:

// IsX reports whether the node has type *X.
func IsX(n Node) bool {
    _, ok := n.(*X)
    return ok
}

// GetX returns the node cast to *X if the node has that type, or nil otherwise.
func GetX(n Node) *X {
    if nn, ok := n.(*X); ok {
        return nn
    }
    return nil
}
}
"""
import re
import os


def put_is(out, typename):
    print >>out, '''
// Is{typename} reports whether the node has type *{typename}.
func Is{typename}(n Node) bool {{
    _, ok := n.(*{typename})
    return ok
}}
'''.format(typename=typename)


def put_get(out, typename):
    print >>out, '''
// Get{typename} returns the node cast to *{typename} if the node has that type, or nil otherwise.
func Get{typename}(n Node) *{typename} {{
    if nn, ok := n.(*{typename}); ok {{
        return nn
    }}
    return nil
}}
'''.format(typename=typename)


def main():
    out = open('boilerplate.go', 'w')
    print >>out, 'package parse'
    for line in file('parse.go'):
        m = re.match(r'^type (.*) struct', line)
        if m:
            in_type = m.group(1)
            put_is(out, in_type)
            put_get(out, in_type)
    out.close()
    os.system('gofmt -w boilerplate.go')


if __name__ == '__main__':
    main()
