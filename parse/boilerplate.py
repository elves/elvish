#!/usr/bin/python2.7
import re


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
func parse{typename}(rd *reader{extraargs}) *{typename} {{
    n := &{typename}{{node: node{{begin: rd.pos}}}}
    n.parse(rd{extranames})
    n.end = rd.pos
    n.sourceText = rd.src[n.begin:n.end]
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
            continue
        m = re.match(
            r'^func \(.* \*(.*)\) parse\(rd \*reader(.*?)\) {$', line)
        if m:
            typename, extraargs = m.groups()
            put_parse(out, typename, extraargs)


if __name__ == '__main__':
    main()
