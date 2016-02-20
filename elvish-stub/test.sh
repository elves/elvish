#!/bin/sh
fail() {
	echo "$*; log left at $log"
	exit 1
}

log=`mktemp elvishXXXXX`

# Start elvish-stub.
elvish-stub > $log &
stub=$!

sleep 0.1

# Wait for startup message.
test `tail -n1 $log` == ok || {
	fail "no startup message or startup too slow"
}

# Send a SIGINT.
kill -2 $stub
ps $stub >/dev/null || {
	fail "stub killed by SIGTERM"
}
test `tail -n1 $log` == 2 || {
	fail "stub didn't record SIGTERM"
}

# Really kill stub.
kill -9 $stub

rm $log
