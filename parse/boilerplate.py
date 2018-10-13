#!/usr/bin/python2.7
"""
Generate helper functions for node types.

For every node type T, it generates the following:

* A IsT func that determines whether a Node is actually of type *T.

* A GetT func that takes Node and returns *T. It examines whether the Node is
  actually of type *T, and if it is, returns it; otherwise it returns nil.

* For each field F of type *[]U, it generates a addToF method that appends a
  node to this field and adds it to the children list.

* For each field F of type *U where U is not a slice, it generates a setF
  method that sets this field and adds it to the children list.

* If the type has a parse method that takes a *paser, it genertes a parseT
  func that takes a *Parser and returns *T. The func creates a new instance of
  *T, sets its begin field, calls its parse method, and set its end and
  sourceText fields.

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

func (n *X) setF(ch *Y) {
    n.F = ch
    addChild(n, ch)
}

func (n *X) addToG(ch *Z) {
    n.G = append(n.G, ch)
    addChild(n, ch)
}

// ParseX parses a node of type *X.
func ParseX(ps *Parser) *X {
    n := &X{node: node{begin: ps.pos}}
    n.parse(ps)
    n.end = ps.pos
    n.sourceText = ps.src[n.begin:n.end]
    return n
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


def put_set(out, parent, field, child):
    print >>out, '''
func (n *{parent}) set{field}(ch *{child}) {{
    n.{field} = ch
    addChild(n, ch)
}}'''.format(parent=parent, field=field, child=child)


def put_addto(out, parent, field, child):
    print >>out, '''
func (n *{parent}) addTo{field}(ch *{child}) {{
    n.{field} = append(n.{field}, ch)
    addChild(n, ch)
}}'''.format(parent=parent, field=field, child=child)


def put_parse(out, typename, extraargs):
    extranames = ', '.join(a.split(' ')[0] for a in extraargs.split(', ')) if extraargs else ''
    print >>out, '''
// Parse{typename} parses a node of type *{typename}.
func Parse{typename}(ps *Parser{extraargs}) *{typename} {{
    n := &{typename}{{node: node{{begin: ps.pos}}}}
    n.parse(ps{extranames})
    n.end = ps.pos
    n.sourceText = ps.src[n.begin:n.end]
    return n
}}'''.format(typename=typename, extraargs=extraargs, extranames=extranames)


def main():
    types = []
    in_type = ''
    out = open('boilerplate.go', 'w')
    print >>out, 'package parse'
    for line in file('parse.go'):
        if in_type:
            if line == '}\n':
                in_type = ''
                continue
            m = re.match(r'^\t(\w+(?:, \w+)*) +(\S+)', line)
            if m:
                fields = m.group(1).split(', ')
                typename = m.group(2)
                if typename.startswith('*'):
                    # Single child
                    [put_set(out, in_type, f, typename[1:]) for f in fields]
                elif typename.startswith('[]*'):
                    # Children list
                    [put_addto(out, in_type, f, typename[3:]) for f in fields]
            continue
        m = re.match(r'^type (.*) struct', line)
        if m:
            in_type = m.group(1)
            put_is(out, in_type)
            put_get(out, in_type)
            continue
        m = re.match(
            r'^func \(.* \*(.*)\) parse\(ps \*Parser(.*?)\) {$', line)
        if m:
            typename, extraargs = m.groups()
            put_parse(out, typename, extraargs)
    out.close()
    os.system('gofmt -w boilerplate.go')


if __name__ == '__main__':
    main()
