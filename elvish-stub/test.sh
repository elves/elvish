#!/bin/sh
fail() {
	echo "$*; log left at $log"
	exit 1
}

lastLogIs() {
    test "$(tail -n1 $log)" = "$*"
}

log=`mktemp elvishXXXXX`

# Start elvish-stub.
elvish-stub > $log &
stub=$!

# Wait for startup message.
for i in `seq 51`; do
    test i == 51 && {
        fail "elvish-stub didn't write startup message within 5 seconds"
    }
    lastLogIs ok && break
    sleep 0.1
done

# Send a SIGINT.
kill -2 $stub
ps $stub >/dev/null || {
	fail "stub killed by SIGTERM"
}
lastLogIs 2 || {
	fail "stub didn't record SIGTERM"
}

# Really kill stub.
kill -9 $stub

rm $log
