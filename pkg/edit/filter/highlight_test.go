package filter_test

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/edit/filter"
	"src.elv.sh/pkg/ui"
)

var highlightTests = []struct {
	name string
	q    string
	want ui.Text
}{
	{
		name: "quoted string",
		q:    `'a'`,
		want: ui.T(`'a'`, ui.FgYellow),
	},
	{
		name: "unsupported primary",
		q:    `$a`,
		want: ui.T(`$a`, ui.FgRed),
	},
	{
		name: "supported list form",
		q:    `[re a]`,
		want: ui.Concat(
			ui.T("[", ui.Bold), ui.T("re", ui.FgGreen),
			ui.T(" a"), ui.T("]", ui.Bold)),
	},
	{
		name: "empty list form",
		q:    `[]`,
		want: ui.T("[]", ui.FgRed),
	},
	{
		name: "unsupported list form",
		q:    `[bad]`,
		want: ui.Concat(
			ui.T("[", ui.Bold), ui.T("bad", ui.FgRed), ui.T("]", ui.Bold)),
	},
	{
		name: "unsupported primary as head of list form",
		q:    `[$a]`,
		want: ui.Concat(
			ui.T("[", ui.Bold), ui.T("$a", ui.FgRed), ui.T("]", ui.Bold)),
	},
}

func TestHighlight(t *testing.T) {
	for _, test := range highlightTests {
		t.Run(test.name, func(t *testing.T) {
			got, _ := filter.Highlight(test.q)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}
