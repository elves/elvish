# Persistent data structure in Go

This is a Go clone of Clojure's persistent data structures.

License is [Eclipse Public License 1.0](http://opensource.org/licenses/eclipse-1.0.php) (like Clojure). See [epl-v10.html](epl-v10.html) for a copy.

API documentation: [![GoDoc](https://godoc.org/github.com/xiaq/persistent?status.png)](https://godoc.org/github.com/xiaq/persistent)

## Implementation notes

The list provided here is a singly-linked list and is very trivial to implement.

The implementation of persistent vector and hash map and based on a series of [excellent](http://blog.higher-order.net/2009/02/01/understanding-clojures-persistentvector-implementation) [blog](http://blog.higher-order.net/2009/09/08/understanding-clojures-persistenthashmap-deftwice) [posts](http://blog.higher-order.net/2010/08/16/assoc-and-clojures-persistenthashmap-part-ii.html) as well as the Clojure source code. Despite the hash map appearing more complicated, the vector is slightly harder to implement due to the "tail array" optimization and some tricky transformation of the tree structure, which is fully replicated here.