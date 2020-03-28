#!/usr/bin/python2.7
import re
import os


def put_compile_s(out, name, intype, extraargs, outtype):
    extranames = ', '.join(a.split(' ')[0] for a in extraargs.split(', ')) if extraargs else ''
    print >>out, '''
func (cp *compiler) {name}Op(n {intype}{extraargs}) {outtype} {{
	return {outtype}{{cp.{name}(n{extranames}), n.Range()}}
}}

func (cp *compiler) {name}Ops(ns []{intype}{extraargs}) []{outtype} {{
	ops := make([]{outtype}, len(ns))
	for i, n := range ns {{
		ops[i] = cp.{name}Op(n{extranames})
	}}
	return ops
}}
'''.format(name=name, intype=intype, outtype=outtype, extraargs=extraargs,
             extranames=extranames)


def main():
    out = open('boilerplate.go', 'w')
    print >>out, '''package eval

import "github.com/elves/elvish/parse"'''
    for fname in 'compile_effect.go', 'compile_value.go':
        for line in file(fname):
            m = re.match(r'^func \(cp \*compiler\) (\w+)\(\w+ ([^,\[\]]+)(.*)\) (\w*Op)Body {$', line)
            if m:
                put_compile_s(out, *m.groups())
    out.close()
    os.system('gofmt -w boilerplate.go')


if __name__ == '__main__':
    main()
