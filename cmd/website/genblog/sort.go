package main

import "sort"

type articleMetas []articleMeta

func less(m1, m2 articleMeta) bool {
	return m1.Timestamp > m2.Timestamp
}

func (a articleMetas) Len() int           { return len(a) }
func (a articleMetas) Less(i, j int) bool { return less(a[i], a[j]) }
func (a articleMetas) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func sortArticleMetas(a []articleMeta) { sort.Sort(articleMetas(a)) }

type articles []article

func (a articles) Len() int           { return len(a) }
func (a articles) Less(i, j int) bool { return less(a[i].articleMeta, a[j].articleMeta) }
func (a articles) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func sortArticles(a articles) { sort.Sort(articles(a)) }
