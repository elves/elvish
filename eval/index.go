package eval

import (
	"errors"
	"fmt"
	"strconv"
)

// Indexer is a Value that can be indexed by Values and yields Values.
type Indexer interface {
	Value
	Index(idx []Value) []Value
}

// IndexOneer is a Value that can be indexed by one Value and yields one Value.
type IndexOneer interface {
	Value
	IndexOne(idx Value) Value
}

// IndexSetter is a Value whose elements can be get as well as set.
type IndexSetter interface {
	IndexOneer
	IndexSet(idx Value, v Value)
}

func mustIndexer(v Value, ec *EvalCtx) Indexer {
	indexer, ok := getIndexer(v, ec)
	if !ok {
		throw(fmt.Errorf("a %s is not indexable", v.Kind()))
	}
	return indexer
}

// getIndexer adapts a Value to an Indexer if there is an adapter. It adapts a
// Caller if ec is not nil.
func getIndexer(v Value, ec *EvalCtx) (Indexer, bool) {
	if indexer, ok := v.(Indexer); ok {
		return indexer, true
	}
	if indexOneer, ok := v.(IndexOneer); ok {
		return IndexOneerIndexer{indexOneer}, true
	}
	if ec != nil {
		if caller, ok := v.(Caller); ok {
			return CallerIndexer{caller, ec}, true
		}
	}
	return nil, false
}

// IndexOneerIndexer adapts an IndexOneer to an Indexer by calling all the
// indicies on the IndexOner and collect the results.
type IndexOneerIndexer struct {
	IndexOneer
}

func (ioi IndexOneerIndexer) Index(vs []Value) []Value {
	results := make([]Value, len(vs))
	for i, v := range vs {
		results[i] = ioi.IndexOneer.IndexOne(v)
	}
	return results
}

func (l List) IndexOne(idx Value) Value {
	i := intIndex(idx)

	if i < 0 {
		i += len(*l.inner)
	}
	if i < 0 || i >= len(*l.inner) {
		throw(ErrIndexOutOfRange)
	}
	return (*l.inner)[i]
}

func (l List) IndexSet(idxv Value, v Value) {
	idx := intIndex(idxv)
	if idx < 0 {
		idx += len(*l.inner)
	}
	if idx < 0 || idx >= len(*l.inner) {
		throw(ErrIndexOutOfRange)
	}
	(*l.inner)[idx] = v
}

func (m Map) IndexOne(idx Value) Value {
	v, ok := (*m.inner)[idx]
	if !ok {
		throw(errors.New("no such key: " + idx.Repr()))
	}
	return v
}

func (m Map) IndexSet(idx Value, v Value) {
	(*m.inner)[idx] = v
}

func intIndex(idx Value) int {
	i, err := strconv.Atoi(ToString(idx))
	if err != nil {
		err := err.(*strconv.NumError)
		if err.Err == strconv.ErrRange {
			throw(ErrIndexOutOfRange)
		} else {
			throw(ErrNeedIntIndex)
		}
	}
	return i
}

// CallerIndexer adapts a Caller to an Indexer.
type CallerIndexer struct {
	Caller
	ec *EvalCtx
}

func (ci CallerIndexer) Index(idx []Value) []Value {
	return captureOutput(ci.ec, func(ec *EvalCtx) {
		ci.Caller.Call(ec, idx)
	})
}
