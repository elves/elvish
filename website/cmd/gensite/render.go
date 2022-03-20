package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"
)

// This file contains functions and types for rendering the site.

// baseDot is the base for all "dot" structures used as the environment of the
// HTML template.
type baseDot struct {
	SiteTitle     string
	Author        string
	RootURL       string
	HomepageTitle string
	Categories    []categoryMeta

	CategoryMap map[string]string
	BaseCSS     string
}

func newBaseDot(bc *siteConf, css string) *baseDot {
	b := &baseDot{bc.Title, bc.Author, bc.RootURL,
		bc.Index.Title, bc.Categories, make(map[string]string), css}
	for _, m := range bc.Categories {
		b.CategoryMap[m.Name] = m.Title
	}
	return b
}

type articleDot struct {
	*baseDot
	article
}

type categoryDot struct {
	*baseDot
	Category string
	Prelude  string
	Groups   []group
	ExtraCSS string
	ExtraJS  string
}

type feedDot struct {
	*baseDot
	Articles     []article
	LastModified rfc3339Time
}

// rfc3339Time wraps time.Time to provide a RFC3339 String() method.
type rfc3339Time time.Time

func (t rfc3339Time) String() string {
	return time.Time(t).Format(time.RFC3339)
}

// contentIs generates a code snippet to fix the free reference "content" in
// the HTML template.
func contentIs(what string) string {
	return fmt.Sprintf(
		`{{ define "content" }} {{ template "%s-content" . }} {{ end }}`,
		what)
}

const fontFaceTemplate = `@font-face { font-family: %v; font-weight: %v; font-style: %v; font-stretch: normal; font-display: block; src: url("%v/fonts/%v.woff2") format("woff");}`

func newTemplate(name, root string, sources ...string) *template.Template {
	t := template.New(name).Funcs(template.FuncMap(map[string]any{
		"is":      func(s string) bool { return s == name },
		"rootURL": func() string { return root },
		"getEnv":  os.Getenv,
		"fontFace": func(family string, weight int, style string, fname string) string {
			return fmt.Sprintf(fontFaceTemplate, family, weight, style, root, fname)
		},
	}))
	for _, source := range sources {
		template.Must(t.Parse(source))
	}
	return t
}

func openForWrite(fname string) *os.File {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func executeToFile(t *template.Template, data any, fname string) {
	file := openForWrite(fname)
	defer file.Close()
	err := t.Execute(file, data)
	if err != nil {
		log.Fatalf("rendering %q: %s", fname, err)
	}
}
