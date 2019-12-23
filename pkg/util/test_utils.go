package util

func mustOK(err error) {
	if err != nil {
		panic(err)
	}
}
