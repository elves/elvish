#!/bin/sh -x

# This test is broken. I haven't yet found a way to start a background process
# with a stdin that I can later write to. Presumably this can be doen with a
# pipe, but there is no way to create pipes explicitly in sh(1).

fail() {
	echo "$*; log left in $dir"
	exit 1
}

waitlog() {
    for i in `seq 51`; do
        test $i = 51 && {
            return 1
        }
        test "$(tail -n1 log)" = "$*" && break
        sleep 0.1
    done
}

dir=`mktemp -d elvishXXXX`
cd "$dir"
mkfifo fifo

# Start elvish-stub.
elvish-stub > log < fifo &
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

rmdir "$dir"
