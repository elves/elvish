package ui_test

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/diag"
	. "src.elv.sh/pkg/ui"
)

var styleRegionsTests = []struct {
	Name     string
	String   string
	Regions  []StylingRegion
	WantText Text
}{
	{
		Name:   "empty string and regions",
		String: "", Regions: nil, WantText: nil,
	},
	{
		Name:   "a single region",
		String: "foobar",
		Regions: []StylingRegion{
			{r(1, 3), FgRed, 0},
		},
		WantText: Concat(T("f"), T("oo", FgRed), T("bar")),
	},

	{
		Name:   "multiple continuos regions",
		String: "foobar",
		Regions: []StylingRegion{
			{r(1, 3), FgRed, 0},
			{r(3, 4), FgGreen, 0},
		},
		WantText: Concat(T("f"), T("oo", FgRed), T("b", FgGreen), T("ar")),
	},

	{
		Name:   "multiple discontinuos regions in wrong order",
		String: "foobar",
		Regions: []StylingRegion{
			{r(4, 5), FgGreen, 0},
			{r(1, 3), FgRed, 0},
		},
		WantText: Concat(T("f"), T("oo", FgRed), T("b"), T("a", FgGreen), T("r")),
	},
	{
		Name:   "regions with the same starting position but differeng priorities",
		String: "foobar",
		Regions: []StylingRegion{
			{r(1, 3), FgRed, 0},
			{r(1, 2), FgGreen, 1},
		},
		WantText: Concat(T("f"), T("o", FgGreen), T("obar")),
	},
	{
		Name:   "overlapping regions with different starting positions",
		String: "foobar",
		Regions: []StylingRegion{
			{r(1, 3), FgRed, 0},
			{r(2, 4), FgGreen, 0},
		},
		WantText: Concat(T("f"), T("oo", FgRed), T("bar")),
	},
}

func r(a, b int) diag.Ranging { return diag.Ranging{From: a, To: b} }

func TestStyleRegions(t *testing.T) {
	for _, test := range styleRegionsTests {
		text := StyleRegions(test.String, test.Regions)
		if !reflect.DeepEqual(text, test.WantText) {
			t.Errorf("got %v, want %v", text, test.WantText)
		}
	}
}
