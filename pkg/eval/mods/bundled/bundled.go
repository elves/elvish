// Package bundled manages modules written in Elvish that are bundled with the
// elvish binary.
package bundled

// Get returns a map of bundled modules.
func Get() map[string]string {
	return map[string]string{
		"epm":              epmElv,
		"readline-binding": readlineBindingElv,
	}
}
