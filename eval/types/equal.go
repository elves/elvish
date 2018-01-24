package types

func Equal(x, y interface{}) bool {
	return x.(Equaler).Equal(y)
}
