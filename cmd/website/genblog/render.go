package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
	"time"
)

// This file contains functions and types for rendering the blog.

// baseDot is the base for all "dot" structures used as the environment of the
// HTML template.
type baseDot struct {
	BlogTitle     string
	Author        string
	RootURL       string
	HomepageTitle string
	Categories    []categoryMeta

	CategoryMap map[string]string
	BaseCSS     string
}

func newBaseDot(bc *blogConf, css string) *baseDot {
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
	Articles []articleMeta
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

func newTemplate(name, root string, sources ...string) *template.Template {
	t := template.New(name).Funcs(template.FuncMap(map[string]interface{}{
		"is":      func(s string) bool { return s == name },
		"rootURL": func() string { return root },
	}))
	for _, source := range sources {
		template.Must(t.Parse(source))
	}
	return t
}

func openForWrite(fname string) (*os.File, error) {
	return os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func executeToFile(t *template.Template, data interface{}, fname string) {
	file, err := openForWrite(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	err = t.Execute(file, data)
	if err != nil {
		log.Fatalf("rendering %q: %s", fname, err)
	}
}
