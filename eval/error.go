package eval

import (
	"../parse"
)

type Error struct {
	node parse.Node
	msg string
}

func (e *Error) Error() string {
	return e.msg
}
