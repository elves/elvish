// Package list implements persistent list.
package list

// List is a persistent list.
type List interface {
	// Len returns the number of values in the list.
	Len() int
	// Conj returns a new list with an additional value in the front.
	Conj(any) List
	// First returns the first value in the list.
	First() any
	// Rest returns the list after the first value.
	Rest() List
}

// Empty is an empty list.
var Empty List = &list{}

type list struct {
	first any
	rest  *list
	count int
}

func (l *list) Len() int {
	return l.count
}

func (l *list) Conj(val any) List {
	return &list{val, l, l.count + 1}
}

func (l *list) First() any {
	return l.first
}

func (l *list) Rest() List {
	return l.rest
}
