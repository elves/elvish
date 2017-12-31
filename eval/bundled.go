package eval

func makeBundled() map[string]string {
	return map[string]string{
		"epm":              epmElv,
		"narrow":           narrowElv,
		"readline-binding": readlineBindingElv,
	}
}
