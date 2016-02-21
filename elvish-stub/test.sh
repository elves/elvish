#!/bin/sh
fail() {
	echo "$*; log left at $log"
	exit 1
}

waitlog() {
    for i in `seq 51`; do
        test $i = 51 && {
            return 1
        }
        test "$(tail -n1 $log)" = "$*" && break
        sleep 0.1
    done
}

log=`mktemp elvishXXXXX`

# Start elvish-stub.
elvish-stub > $log &
stub=$!

# Wait for startup message.
waitlog ok || fail "didn't write startup message"

# Send SIG{INT,TERM,TSTP}
for sig in 2 15 20; do
    kill -$sig $stub
    ps $stub >/dev/null || {
        fail "stub killed by signal #$sig"
    }
    waitlog $sig || fail "didn't record signal #$sig"
done

# Really kill stub.
kill -9 $stub

rm $log
