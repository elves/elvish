package ui

import (
	"sort"

	"src.elv.sh/pkg/diag"
)

// StylingRegion represents a region to apply styling.
type StylingRegion struct {
	diag.Ranging
	Styling  Styling
	Priority int
}

// StyleRegions applies styling to the specified regions in s.
//
// The regions are sorted by start position. If multiple Regions share the same
// starting position, the one with the highest priority is kept; the other
// regions are removed. If a Region starts before the end of the previous
// Region, it is also removed.
func StyleRegions(s string, regions []StylingRegion) Text {
	regions = fixRegions(regions)

	var text Text
	lastTo := 0
	for _, r := range regions {
		if r.From > lastTo {
			// Add text between regions or before the first region.
			text = append(text, &Segment{Text: s[lastTo:r.From]})
		}
		text = append(text,
			StyleSegment(&Segment{Text: s[r.From:r.To]}, r.Styling))
		lastTo = r.To
	}
	if len(s) > lastTo {
		// Add text after the last region.
		text = append(text, &Segment{Text: s[lastTo:]})
	}
	return text
}

func fixRegions(regions []StylingRegion) []StylingRegion {
	regions = append([]StylingRegion(nil), regions...)
	// Sort regions by their start positions. Regions with the same start
	// position are sorted by decreasing priority.
	sort.Slice(regions, func(i, j int) bool {
		a, b := regions[i], regions[j]
		return a.From < b.From || (a.From == b.From && a.Priority > b.Priority)
	})
	// Remove overlapping regions, preferring the ones that appear earlier.
	var newRegions []StylingRegion
	lastTo := 0
	for _, r := range regions {
		if r.From < lastTo {
			// Overlaps with the last one
			continue
		}
		newRegions = append(newRegions, r)
		lastTo = r.To
	}
	return newRegions
}
