// Package list implements persistent list.
package list

// List implements a persistent list. The empty value is a valid empty list.
type List struct {
	first interface{}
	rest  *List
	count int
}

// Empty is an empty List.
var Empty = &List{}

// Cons returns a new List with an additional value in the front.
func (l *List) Cons(val interface{}) *List {
	return &List{val, l, l.count + 1}
}

// First returns the first value in the list.
func (l *List) First() interface{} {
	return l.first
}

// Rest returns the list after the first value.
func (l *List) Rest() *List {
	return l.rest
}

// Count returns the number of values in the list.
func (l *List) Count() int {
	return l.count
}
