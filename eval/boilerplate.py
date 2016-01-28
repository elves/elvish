#!/usr/bin/python2.7
import re


def put_compile_s(out, name, intype, outtype):
    print >>out, '''
func (cp *compiler) {name}s(ns []{intype}) []{outtype} {{
    ops := make([]{outtype}, len(ns))
    for i, n := range ns {{
        ops[i] = cp.{name}(n)
    }}
    return ops
}}'''.format(name=name, intype=intype, outtype=outtype)


def main():
    out = open('boilerplate.go', 'w')
    print >>out, '''package eval
import "github.com/elves/elvish/parse"'''
    for line in file('compile.go'):
        m = re.match(r'^func \(cp \*compiler\) (\w+)\(\w+ (.+)\) (\w*[oO]p) {$', line)
        if m:
            put_compile_s(out, *m.groups())


if __name__ == '__main__':
    main()
