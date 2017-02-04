#!/bin/bash

out=./builtin-modules.go

{
	echo "package eval"
	echo "var builtinModules = map[string]string{"

	for f in *.elv; do
		echo -n "\"${f%.elv}\": \`"
		cat $f | sed 's/`/``/g'
		echo '`,'
	done

	echo "}"
} > $out

gofmt -w $out
