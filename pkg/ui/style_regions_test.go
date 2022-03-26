package ui_test

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/ui"
)

var styleRegionsTests = []struct {
	Name     string
	String   string
	Regions  []ui.StylingRegion
	WantText ui.Text
}{
	{
		Name:   "empty string and regions",
		String: "", Regions: nil, WantText: nil,
	},
	{
		Name:   "a single region",
		String: "foobar",
		Regions: []ui.StylingRegion{
			{r(1, 3), ui.FgRed, 0},
		},
		WantText: ui.Concat(ui.T("f"), ui.T("oo", ui.FgRed), ui.T("bar")),
	},

	{
		Name:   "multiple continuous regions",
		String: "foobar",
		Regions: []ui.StylingRegion{
			{r(1, 3), ui.FgRed, 0},
			{r(3, 4), ui.FgGreen, 0},
		},
		WantText: ui.Concat(ui.T("f"), ui.T("oo", ui.FgRed), ui.T("b", ui.FgGreen), ui.T("ar")),
	},

	{
		Name:   "multiple discontinuous regions in wrong order",
		String: "foobar",
		Regions: []ui.StylingRegion{
			{r(4, 5), ui.FgGreen, 0},
			{r(1, 3), ui.FgRed, 0},
		},
		WantText: ui.Concat(ui.T("f"), ui.T("oo", ui.FgRed), ui.T("b"), ui.T("a", ui.FgGreen), ui.T("r")),
	},
	{
		Name:   "regions with the same starting position but differeng priorities",
		String: "foobar",
		Regions: []ui.StylingRegion{
			{r(1, 3), ui.FgRed, 0},
			{r(1, 2), ui.FgGreen, 1},
		},
		WantText: ui.Concat(ui.T("f"), ui.T("o", ui.FgGreen), ui.T("obar")),
	},
	{
		Name:   "overlapping regions with different starting positions",
		String: "foobar",
		Regions: []ui.StylingRegion{
			{r(1, 3), ui.FgRed, 0},
			{r(2, 4), ui.FgGreen, 0},
		},
		WantText: ui.Concat(ui.T("f"), ui.T("oo", ui.FgRed), ui.T("bar")),
	},
}

func r(a, b int) diag.Ranging { return diag.Ranging{From: a, To: b} }

func TestStyleRegions(t *testing.T) {
	for _, test := range styleRegionsTests {
		text := ui.StyleRegions(test.String, test.Regions)
		if !reflect.DeepEqual(text, test.WantText) {
			t.Errorf("got %v, want %v", text, test.WantText)
		}
	}
}
