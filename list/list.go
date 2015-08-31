// Package list implements persistent list.
package list

// List implements a persistent list. The empty value is a valid empty list.
type List struct {
	first interface{}
	rest  *List
	count int
}

var Empty = &List{}

func (l *List) Cons(val interface{}) *List {
	return &List{val, l, l.count + 1}
}

func (l *List) First() interface{} {
	return l.first
}

func (l *List) Rest() *List {
	return l.rest
}

func (l *List) Count() int {
	return l.count
}
