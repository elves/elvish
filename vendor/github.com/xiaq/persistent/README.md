# Persistent data structure in Go

[![GoDoc](https://godoc.org/github.com/xiaq/persistent?status.svg)](https://godoc.org/github.com/xiaq/persistent)
[![Build Status](https://travis-ci.org/xiaq/persistent.svg?branch=master)](https://travis-ci.org/xiaq/persistent)
[![Coverage Status](https://coveralls.io/repos/github/xiaq/persistent/badge.svg?branch=master)](https://coveralls.io/github/xiaq/persistent?branch=master)

This is a Go clone of Clojure's persistent data structures.

The API is not stable yet. **DO NOT USE** unless you are willing to cope with
API changes.

License is [Eclipse Public License 1.0](http://opensource.org/licenses/eclipse-1.0.php) (like Clojure). See [epl-v10.html](epl-v10.html) for a copy.


## Implementation notes

The list provided here is a singly-linked list and is very trivial to implement.

The implementation of persistent vector and hash map and based on a series of [excellent](http://blog.higher-order.net/2009/02/01/understanding-clojures-persistentvector-implementation) [blog](http://blog.higher-order.net/2009/09/08/understanding-clojures-persistenthashmap-deftwice) [posts](http://blog.higher-order.net/2010/08/16/assoc-and-clojures-persistenthashmap-part-ii.html) as well as the Clojure source code. Despite the hash map appearing more complicated, the vector is slightly harder to implement due to the "tail array" optimization and some tricky transformation of the tree structure, which is fully replicated here.

## Benchmarking results

### Vectors

Compared to native slices,

*   Adding elements is anywhere from 2x to 8x as slow.

*   Sequential read is about 9x as slow.

*   Random read is about 7x as slow.

Benchmarked on an early 2015 MacBook Pro, with Go 1.9:

```
goos: darwin
goarch: amd64
pkg: github.com/xiaq/persistent/vector
BenchmarkConsNativeN1-4                  1000000              2457 ns/op
BenchmarkConsNativeN2-4                   300000              4418 ns/op
BenchmarkConsNativeN3-4                    30000             55424 ns/op
BenchmarkConsNativeN4-4                      300           4493289 ns/op
BenchmarkConsPersistentN1-4               100000             12250 ns/op 4.99x
BenchmarkConsPersistentN2-4                50000             26394 ns/op 5.97x
BenchmarkConsPersistentN3-4                 3000            452146 ns/op 8.16x
BenchmarkConsPersistentN4-4                  100          13057887 ns/op 2.91x
BenchmarkNthSeqNativeN4-4                  30000             43156 ns/op
BenchmarkNthSeqPersistentN4-4               3000            399193 ns/op 9.25x
BenchmarkNthRandNative-4                   20000             73860 ns/op
BenchmarkNthRandPersistent-4                3000            546124 ns/op 7.39x
BenchmarkEqualNative-4                     50000             23828 ns/op
BenchmarkEqualPersistent-4                  2000           1020893 ns/op 42.84x
```

### Hash map

Compared to native maps, adding elements is about 3-6x slow. Difference is
more pronunced when keys are sequential integers, but that workload is very
rare in the real world.

Benchmarked on an early 2015 MacBook Pro, with Go 1.9:

```
goos: darwin
goarch: amd64
pkg: github.com/xiaq/persistent/hashmap
BenchmarkSequentialConsNative1-4                  300000              4143 ns/op
BenchmarkSequentialConsNative2-4                   10000            130423 ns/op
BenchmarkSequentialConsNative3-4                     300           4600842 ns/op
BenchmarkSequentialConsPersistent1-4              100000             14005 ns/op 3.38x
BenchmarkSequentialConsPersistent2-4                2000            641820 ns/op 4.92x
BenchmarkSequentialConsPersistent3-4                  20          55180306 ns/op 11.99x
BenchmarkRandomStringsConsNative1-4               200000              7536 ns/op
BenchmarkRandomStringsConsNative2-4                 5000            264489 ns/op
BenchmarkRandomStringsConsNative3-4                  100          12132244 ns/op
BenchmarkRandomStringsConsPersistent1-4            50000             29109 ns/op 3.86x
BenchmarkRandomStringsConsPersistent2-4             1000           1327321 ns/op 5.02x
BenchmarkRandomStringsConsPersistent3-4               20          74204196 ns/op 6.12x
```
