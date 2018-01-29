package vartypes

type callback struct {
	set func(interface{}) error
	get func() interface{}
}

// NewCallback makes a variable from a set callback and a get callback.
func NewCallback(set func(interface{}) error, get func() interface{}) Variable {
	return &callback{set, get}
}

func (cv *callback) Set(val interface{}) error {
	return cv.set(val)
}

func (cv *callback) Get() interface{} {
	return cv.get()
}

type roCallback func() interface{}

// NewRoCallback makes a read-only variable from a get callback.
func NewRoCallback(get func() interface{}) Variable {
	return roCallback(get)
}

func (cv roCallback) Set(interface{}) error {
	return errRoCannotBeSet
}

func (cv roCallback) Get() interface{} {
	return cv()
}
