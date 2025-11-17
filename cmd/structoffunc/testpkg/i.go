package testpkg

import (
	myjson "encoding/json"
	"os"
)

type I interface {
	Ident() X
	Selector() os.File
	Star() *int
	Array() [10]int
	Slice() []int
	Struct() struct {
		A    int
		B, C string
	}
	Func() func(a int) string
	Interface() interface {
		X()
		Y(int) string
	}
	Map() map[string]int
	Chan() chan int
	ChanSend() chan<- int
	ChanRecv() <-chan int

	AliasedModule() myjson.Number
	NoReturn()
	NamedParams(a int, b string)
	UnnamedParams(int, string)
	MultipleParamsWithSameType(a, b int, c string)
	Variadic(a ...int)
}

type X int
