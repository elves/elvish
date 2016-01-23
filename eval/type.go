package eval

type Type interface {
	String() string
}

type stringType struct {
}

func (st stringType) String() string {
	return "string"
}

type boolType struct {
}

func (bt boolType) String() string {
	return "bool"
}

type exitusType struct {
}

func (et exitusType) String() string {
	return "exitus"
}

type tableType struct {
}

func (tt tableType) String() string {
	return "table"
}

type callableType struct {
}

func (ct callableType) String() string {
	return "callable"
}

type ratType struct{}

func (rt ratType) String() string {
	return "rat"
}

var typenames = map[string]Type{
	"string":   stringType{},
	"exitus":   exitusType{},
	"bool":     boolType{},
	"table":    tableType{},
	"callable": callableType{},
	"rat":      ratType{},
}
