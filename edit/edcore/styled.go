package edcore

import (
	"errors"

	"github.com/elves/elvish/edit/ui"
	"github.com/xiaq/persistent/vector"
)

var errStyledStyles = errors.New("styles must either be a string or list of strings")

// A constructor for *ui.Styled, for use in Elvish script.
func styled(text string, styles interface{}) (*ui.Styled, error) {
	switch styles := styles.(type) {
	case string:
		return &ui.Styled{text, ui.StylesFromString(styles)}, nil
	case vector.Vector:
		converted := make([]string, 0, styles.Len())
		for it := styles.Iterator(); it.HasElem(); it.Next() {
			elem, ok := it.Elem().(string)
			if !ok {
				return nil, errStyledStyles
			}
			converted = append(converted, elem)
		}
		return &ui.Styled{text, ui.Styles(converted)}, nil
	default:
		return nil, errStyledStyles
	}
}
