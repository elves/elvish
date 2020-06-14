package main

//go:generate ./gen-include

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatal("Usage: genblog <src dir> <dst dir>")
	}
	srcDir, dstDir := args[0], args[1]
	srcFile := func(elem ...string) string {
		elem = append([]string{srcDir}, elem...)
		return filepath.Join(elem...)
	}
	dstFile := func(elem ...string) string {
		elem = append([]string{dstDir}, elem...)
		return filepath.Join(elem...)
	}

	// Read blog configuration.
	conf := &blogConf{}
	decodeTOML(srcFile("index.toml"), conf)
	if conf.RootURL == "" {
		log.Fatal("RootURL must be specified; needed by feed and sitemap")
	}
	if conf.Template == "" {
		log.Fatal("Template must be specified")
	}
	if conf.BaseCSS == nil {
		log.Fatal("BaseCSS must be specified")
	}

	template := readFile(srcFile(conf.Template))
	baseCSS := catInDir(srcDir, conf.BaseCSS)

	// Initialize templates. They are all initialized from the same source code,
	// plus a snippet to fix the "content" reference.
	categoryTmpl := newTemplate("category", "..", template, contentIs("category"))
	articleTmpl := newTemplate("article", "..", template, contentIs("article"))
	homepageTmpl := newTemplate("homepage", ".", template, contentIs("article"))
	feedTmpl := newTemplate("feed", ".", feedTemplText)

	// Base for the {{ . }} object used in all templates.
	base := newBaseDot(conf, baseCSS)

	// Up to conf.FeedPosts recent posts, used in the feed.
	recents := recentArticles{nil, conf.FeedPosts}
	// Last modified time of the newest post, used in the feed.
	var lastModified time.Time

	// Whether the "all" category has been requested.
	hasAllCategory := false
	// Meta of all articles, used to generate the index of the "all", if if is
	// requested.
	allArticleMetas := []articleMeta{}

	// Paths of all generated URLs, relative to the destination directory,
	// always without "index.html". Used to generate the sitemap.
	allPaths := []string{""}

	// Render a category index.
	renderCategoryIndex := func(name, prelude, css, js string, articles []articleMeta) {
		// Add category index to the sitemap, without "/index.html"
		allPaths = append(allPaths, name)
		// Create directory
		catDir := dstFile(name)
		err := os.MkdirAll(catDir, 0755)
		if err != nil {
			log.Fatal(err)
		}

		// Generate index
		cd := &categoryDot{base, name, prelude, articles, css, js}
		executeToFile(categoryTmpl, cd, filepath.Join(catDir, "index.html"))
	}

	for _, cat := range conf.Categories {
		if cat.Name == "all" {
			// The "all" category has been requested. It is a pseudo-category in
			// that it doesn't need to have any associated category
			// configuration file. We cannot render the category index now
			// because we haven't seen all articles yet. Render it later.
			hasAllCategory = true
			continue
		}

		catConf := &categoryConf{}
		decodeTOML(srcFile(cat.Name, "index.toml"), catConf)

		prelude := ""
		if catConf.Prelude != "" {
			prelude = readFile(srcFile(cat.Name, catConf.Prelude+".html"))
		}
		css := catInDir(srcFile(cat.Name), catConf.ExtraCSS)
		js := catInDir(srcFile(cat.Name), catConf.ExtraJS)
		renderCategoryIndex(cat.Name, prelude, css, js, catConf.Articles)

		// Generate articles
		for _, am := range catConf.Articles {
			// Add article URL to sitemap.
			p := filepath.Join(cat.Name, am.Name+".html")
			allPaths = append(allPaths, p)

			a := getArticle(article{Category: cat.Name}, am, srcFile(cat.Name))
			modTime := time.Time(a.LastModified)
			if modTime.After(lastModified) {
				lastModified = modTime
			}

			// Generate article page.
			ad := &articleDot{base, a}
			executeToFile(articleTmpl, ad, dstFile(p))

			allArticleMetas = append(allArticleMetas, a.articleMeta)
			recents.insert(a)
		}
	}

	// Generate "all category"
	if hasAllCategory {
		sort.Slice(allArticleMetas, func(i, j int) bool {
			return allArticleMetas[i].Timestamp > allArticleMetas[j].Timestamp
		})
		renderCategoryIndex("all", "", "", "", allArticleMetas)
	}

	// Generate index page. XXX(xiaq): duplicated code with generating ordinary
	// article pages.
	a := getArticle(article{IsHomepage: true, Category: "homepage"}, conf.Index, srcDir)
	ad := &articleDot{base, a}
	executeToFile(homepageTmpl, ad, dstFile("index.html"))

	// Generate feed.
	feedArticles := recents.articles
	fd := feedDot{base, feedArticles, rfc3339Time(lastModified)}
	executeToFile(feedTmpl, fd, dstFile("feed.atom"))

	// Generate site map.
	file := openForWrite(dstFile("sitemap.txt"))
	defer file.Close()
	for _, p := range allPaths {
		fmt.Fprintf(file, "%s/%s\n", conf.RootURL, p)
	}
}
