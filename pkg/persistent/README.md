# Persistent data structure in Go

This is a Go clone of Clojure's persistent data structures.

License is
[Eclipse Public License 1.0](http://opensource.org/licenses/eclipse-1.0.php)
(like Clojure).

## Implementation notes

The list provided here is a singly-linked list and is very trivial to implement.

The implementation of persistent vector and hash map and based on a series of
[excellent](http://blog.higher-order.net/2009/02/01/understanding-clojures-persistentvector-implementation)
[blog](http://blog.higher-order.net/2009/09/08/understanding-clojures-persistenthashmap-deftwice)
[posts](http://blog.higher-order.net/2010/08/16/assoc-and-clojures-persistenthashmap-part-ii.html)
as well as the Clojure source code. Despite the hash map appearing more
complicated, the vector is slightly harder to implement due to the "tail array"
optimization and some tricky transformation of the tree structure, which is
fully replicated here.

## Benchmarking results

### Vectors

Compared to native slices,

-   Adding elements is anywhere from 5x to 9x as slow.

-   Read (sequential or random) is about 6x as slow.

Benchmarked on an MacBook Air (M1, 2020), with Go 1.17.5:

```
BenchmarkConjNativeN1-8              1779234           673.3 ns/op
BenchmarkConjNativeN2-8               948654          1220 ns/op
BenchmarkConjNativeN3-8                61242         20138 ns/op
BenchmarkConjNativeN4-8                 1222        968176 ns/op
BenchmarkConjPersistentN1-8           264488          4462 ns/op 6.63x
BenchmarkConjPersistentN2-8           119526          9885 ns/op 8.10x
BenchmarkConjPersistentN3-8             6760        173995 ns/op 8.64x
BenchmarkConjPersistentN4-8              212       5576977 ns/op 5.76x
BenchmarkIndexSeqNativeN4-8            32031         37344 ns/op
BenchmarkIndexSeqPersistentN4-8         6145        192151 ns/op 5.15x
BenchmarkIndexRandNative-8             31366         38016 ns/op
BenchmarkIndexRandPersistent-8          5434        216284 ns/op 5.69x
BenchmarkEqualNative-8                110090         10738 ns/op
BenchmarkEqualPersistent-8              2121        557334 ns/op 51.90x
```

### Hash map

Compared to native maps, adding elements is about 3-6x slow. Difference is more
pronunced when keys are sequential integers, but that workload is very rare in
the real world.

Benchmarked on an MacBook Air (M1, 2020), with Go 1.17.5:

```
goos: darwin
goarch: arm64
pkg: src.elv.sh/pkg/persistent/hashmap
BenchmarkSequentialConjNative1-8              620540          1900 ns/op
BenchmarkSequentialConjNative2-8               22918         52209 ns/op
BenchmarkSequentialConjNative3-8                 567       2115886 ns/op
BenchmarkSequentialConjPersistent1-8          169776          7026 ns/op 3.70x
BenchmarkSequentialConjPersistent2-8            3374        354031 ns/op 6.78x
BenchmarkSequentialConjPersistent3-8              51      23091870 ns/op 10.91x
BenchmarkRandomStringsConjNative1-8           379147          3155 ns/op
BenchmarkRandomStringsConjNative2-8            10000        117332 ns/op
BenchmarkRandomStringsConjNative3-8              292       4034937 ns/op
BenchmarkRandomStringsConjPersistent1-8        96504         12207 ns/op 3.87x
BenchmarkRandomStringsConjPersistent2-8         1910        615644 ns/op 5.25x
BenchmarkRandomStringsConjPersistent3-8           33      31928604 ns/op 7.91x
```
