package eval

import (
	"errors"
	"strconv"
)

// Indexer is a Value that can be indexed by a Value and yields a Value.
type Indexer interface {
	Value
	Index(idx Value) Value
}

// IndexSetter is a Value whose elements can be get as well as set.
type IndexSetter interface {
	Indexer
	IndexSet(idx Value, v Value)
}

func (l List) Index(idx Value) Value {
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

func (m Map) Index(idx Value) Value {
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
