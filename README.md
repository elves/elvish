# Persistent data structure in Go

[![GoDoc](https://godoc.org/github.com/xiaq/persistent?status.svg)](https://godoc.org/github.com/xiaq/persistent)
[![Build Status](https://travis-ci.org/xiaq/persistent.svg?branch=master)](https://travis-ci.org/xiaq/persistent)
[![Coverage Status](https://coveralls.io/repos/github/xiaq/persistent/badge.svg?branch=master)](https://coveralls.io/github/xiaq/persistent?branch=master)

This is a Go clone of Clojure's persistent data structures.

License is [Eclipse Public License 1.0](http://opensource.org/licenses/eclipse-1.0.php) (like Clojure). See [epl-v10.html](epl-v10.html) for a copy.


## Implementation notes

The list provided here is a singly-linked list and is very trivial to implement.

The implementation of persistent vector and hash map and based on a series of [excellent](http://blog.higher-order.net/2009/02/01/understanding-clojures-persistentvector-implementation) [blog](http://blog.higher-order.net/2009/09/08/understanding-clojures-persistenthashmap-deftwice) [posts](http://blog.higher-order.net/2010/08/16/assoc-and-clojures-persistenthashmap-part-ii.html) as well as the Clojure source code. Despite the hash map appearing more complicated, the vector is slightly harder to implement due to the "tail array" optimization and some tricky transformation of the tree structure, which is fully replicated here.

## Benchmarking results

### Vectors

Adding elements to a persistent vector is about 4x as slow as appending elements to a native slices for small-ish data (hundreds of elements). The relative performance **improve** with the size of the vector, perhaps due to the fact that adding elements to the persistent vector never requires reallocating the entire vector, while appending elements to a native slice can do that.

These are run on a MacBook Pro, early 2016 model.

Run 1:

```
BenchmarkNativeAppendN1-4        1000000              1993 ns/op
BenchmarkNativeAppendN2-4         300000              3495 ns/op
BenchmarkNativeAppendN3-4          30000             43821 ns/op
BenchmarkNativeAppendN4-4            500           3208284 ns/op
BenchmarkConsN1-4                 200000              7711 ns/op 3.87x
BenchmarkConsN2-4                 100000             15974 ns/op 4.57x
BenchmarkConsN3-4                   5000            253898 ns/op 5.79x
BenchmarkConsN4-4                    200           8744860 ns/op 2.73x
```

Run 2:

```
BenchmarkNativeAppendN1-4         500000              2209 ns/op
BenchmarkNativeAppendN2-4         300000              5998 ns/op
BenchmarkNativeAppendN3-4          30000             46853 ns/op
BenchmarkNativeAppendN4-4            500           2790387 ns/op
BenchmarkConsN1-4                 200000             10594 ns/op 4.80x
BenchmarkConsN2-4                 100000             26767 ns/op 4.46x
BenchmarkConsN3-4                   5000            373414 ns/op 7.97x
BenchmarkConsN4-4                    100          10536014 ns/op 3.78x
```

### Hash map

Adding elements to a persistent hash map is about 5x as slow as adding elements to a native map for small-ish data (hundreds of elements). The relative performance **worsens** with the size of the map.

These are run on a $5/mo DigitalOcean VPS.

```
BenchmarkNativeSequentialAddN1    200000              6855 ns/op
BenchmarkNativeSequentialAddN2     10000            217681 ns/op
BenchmarkNativeSequentialAddN3       200           7039884 ns/op
BenchmarkSequentialConsN1          50000             28746 ns/op 4.2x
BenchmarkSequentialConsN2           1000           1208015 ns/op 5.5x
BenchmarkSequentialConsN3             10         138323217 ns/op 19.6x
BenchmarkNativeRandomStringsAddN1         100000             12900 ns/op
BenchmarkNativeRandomStringsAddN2           3000            462787 ns/op
BenchmarkNativeRandomStringsAddN3             50          26765594 ns/op
BenchmarkRandomStringsConsN1               30000             46482 ns/op 3.6x
BenchmarkRandomStringsConsN2                1000           2503646 ns/op 5.4x
BenchmarkRandomStringsConsN3                  10         200831466 ns/op 7.5x
```
