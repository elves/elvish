package edcore

import (
	"strings"

	"github.com/elves/elvish/eval"
	styled2 "github.com/elves/elvish/styled"
	"github.com/xiaq/persistent/vector"
)

func styled(fm *eval.Frame, text string, styles interface{}) (*styled2.Text, error) {
	transformers := make([]interface{}, 0)

	switch styles := styles.(type) {
	case string:
		for _, s := range strings.Split(styles, ";") {
			transformers = append(transformers, s)
		}
	case vector.Vector:
		for it := styles.Iterator(); it.HasElem(); it.Next() {
			transformers = append(transformers, it.Elem())
		}
	default:
		return eval.StyledBuiltin(fm, text, styles)
	}

	return eval.StyledBuiltin(fm, text, transformers...)
}
