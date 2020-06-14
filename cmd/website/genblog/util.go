package main

import "sort"

func sortArticleMetas(a []articleMeta) {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Timestamp > a[j].Timestamp
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
