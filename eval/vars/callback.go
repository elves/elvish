package vars

type callback struct {
	set func(interface{}) error
	get func() interface{}
}

// FromSetGet makes a variable from a set callback and a get callback.
func FromSetGet(set func(interface{}) error, get func() interface{}) Var {
	return &callback{set, get}
}

func (cv *callback) Set(val interface{}) error {
	return cv.set(val)
}

func (cv *callback) Get() interface{} {
	return cv.get()
}

type roCallback func() interface{}

// FromGet makes a variable from a get callback. The variable is read-only.
func FromGet(get func() interface{}) Var {
	return roCallback(get)
}

func (cv roCallback) Set(interface{}) error {
	return errSetReadOnlyVar
}

func (cv roCallback) Get() interface{} {
	return cv()
}
