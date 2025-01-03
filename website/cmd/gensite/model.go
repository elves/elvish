package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// This file contains functions and types for parsing and manipulating the
// in-memory representation of the site.

// siteConf represents the global site configuration.
type siteConf struct {
	Title      string
	Author     string
	Categories []categoryMeta
	Index      articleMeta
	FeedPosts  int
	RootURL    string
	Template   string
	BaseCSS    []string
}

// categoryMeta represents the metadata of a cateogory, found in the global
// site configuration.
type categoryMeta struct {
	Name    string
	Title   string
	NavHTML template.HTML
}

// categoryConf represents the configuration of a category. Note that the
// metadata is found in the global site configuration and not duplicated here.
type categoryConf struct {
	Prelude   string
	AutoIndex bool
	ExtraCSS  []string
	ExtraJS   []string
	Articles  []articleMeta
	Groups    []groupMeta
}

// articleMeta represents the metadata of an article, found in a category
// configuration.
type articleMeta struct {
	Name      string
	Title     string
	Note      string
	Group     int
	Timestamp string
	ExtraCSS  []string
	ExtraJS   []string
}

// article represents an article, including all information that is needed to
// render it.
type article struct {
	articleMeta
	IsHomepage   bool
	Category     string
	Content      string
	ExtraCSS     string
	ExtraJS      string
	LastModified rfc3339Time
}

// Metadata of a group, found in a category index.toml.
type groupMeta struct {
	Intro string
}

// All information about a group to render it.
type group struct {
	groupMeta
	Articles []articleMeta
}

type recentArticles struct {
	articles []article
	max      int
}

func (ra *recentArticles) insert(a article) {
	// Find a place to insert.
	var i int
	for i = len(ra.articles); i > 0; i-- {
		if ra.articles[i-1].Timestamp > a.Timestamp {
			break
		}
	}
	// If we are at the end, insert only if we haven't reached the maximum
	// number of articles.
	if i == len(ra.articles) {
		if i < ra.max {
			ra.articles = append(ra.articles, a)
		}
		return
	}
	// If not, make space and insert.
	if len(ra.articles) < ra.max {
		ra.articles = append(ra.articles, article{})
	}
	copy(ra.articles[i+1:], ra.articles[i:])
	ra.articles[i] = a
}

// decodeTOML decodes the named file in TOML into a pointer.
func decodeTOML(fname string, v any) {
	_, err := toml.DecodeFile(fname, v)
	if err != nil {
		log.Fatalln(err)
	}
}

func readFile(fname string) string {
	content, err := os.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func catInDir(dirname string, fnames []string) string {
	var sb strings.Builder
	for _, fname := range fnames {
		sb.WriteString(readFile(filepath.Join(dirname, fname)))
	}
	return sb.String()
}

func getArticle(a article, am articleMeta, dir string) article {
	fname := filepath.Join(dir, am.Name+".html")
	content := readFile(fname)
	fileInfo, err := os.Stat(fname)
	if err != nil {
		log.Fatal(err)
	}
	modTime := fileInfo.ModTime()
	css := catInDir(dir, am.ExtraCSS)
	js := catInDir(dir, am.ExtraJS)
	return article{
		am, a.IsHomepage, a.Category, content, css, js, rfc3339Time(modTime)}
}
