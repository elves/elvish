// Package list implements persistent list.
package list

import "github.com/xiaq/persistent/types"

// List is a persistent list.
type List interface {
	types.Equaler
	// Len returns the number of values in the list.
	Len() int
	// Cons returns a new list with an additional value in the front.
	Cons(interface{}) List
	// First returns the first value in the list.
	First() interface{}
	// Rest returns the list after the first value.
	Rest() List
}

// Empty is an empty list.
var Empty List = &list{}

type list struct {
	first interface{}
	rest  *list
	count int
}

func (l *list) Len() int {
	return l.count
}

func (l *list) Cons(val interface{}) List {
	return &list{val, l, l.count + 1}
}

func (l *list) First() interface{} {
	return l.first
}

func (l *list) Rest() List {
	return l.rest
}

func (l *list) Equal(other interface{}) bool {
	l2, ok := other.(List)
	if !ok {
		return false
	}
	return Equal(l, l2)
}

// Equal returns whether two List values are equal to each other. The values are
// compared using the types.Equaler interface if the value in l1 implements
// types.Equaler, or with == otherwise.
func Equal(l1, l2 List) bool {
	if l1.Len() != l2.Len() {
		return false
	}
	for i := 0; i < l1.Len(); i++ {
		v1 := l1.First()
		v2 := l2.First()
		if v1eq, ok := v1.(types.Equaler); ok {
			if !v1eq.Equal(v2) {
				return false
			}
		} else {
			if v1 != v2 {
				return false
			}
		}
		l1 = l1.Rest()
		l2 = l2.Rest()
	}
	return true
}
