package types

func Hash(v interface{}) uint32 {
	return v.(Hasher).Hash()
}
