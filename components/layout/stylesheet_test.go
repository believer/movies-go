package layout

import (
	"context"
	"io"
	"testing"
	"testing/fstest"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func renderStylesheetHref(t *testing.T) string {
	t.Helper()

	r, w := io.Pipe()

	go func() {
		_ = Html().Render(context.Background(), w)
		_ = w.Close()
	}()

	doc, err := goquery.NewDocumentFromReader(r)
	assert.NoError(t, err, "Failed to read template")

	href, _ := doc.Find(`link[rel="stylesheet"]`).Attr("href")

	return href
}

func TestStylesheet(t *testing.T) {
	prev := stylesheetURL
	t.Cleanup(func() { stylesheetURL = prev })

	assert.Equal(t, "/styles.css", renderStylesheetHref(t))

	// SHA-256 of "body{}" starts with 7c9804, of "body{color:red}" with 15c42a
	VersionStylesheet([]byte("body{}"))
	assert.Equal(t, "/styles.css?v=7c9804", renderStylesheetHref(t))

	VersionStylesheet([]byte("body{color:red}"))
	assert.Equal(t, "/styles.css?v=15c42a", renderStylesheetHref(t))
}

func TestBuildStylesheet(t *testing.T) {
	file := func(css string) *fstest.MapFile {
		return &fstest.MapFile{Data: []byte(css)}
	}

	// Files follow the directory tree: components/b/a.css sorts first by
	// file name, and components/a-b sorts before components/a/ by path, so
	// either sort would order these differently
	css, err := BuildStylesheet(fstest.MapFS{
		"views/stats.css":            file(".stats{}\n"),
		"views/feed.css":             file(".feed{}\n"),
		"components/b/a.css":         file(".b-a{}\n"),
		"components/a/z.css":         file(".a-z{}\n"),
		"components/a/a.templ":       file("templ A() {}\n"),
		"styles/tokens.css":          file(":root{}\n"),
		"styles/base.css":            file("@layer base{}\n"),
		"components/a/nested/x.css":  file(".a-nested-x{}\n"),
		"components/ab/feature.css":  file(".ab{}\n"),
		"components/a-b/feature.css": file(".a-b{}\n"),
	})

	assert.NoError(t, err)
	assert.Equal(t, `@layer base, components, views;

/* styles/base.css */
@layer base{}

/* styles/tokens.css */
:root{}

@layer components {

/* components/a/nested/x.css */
.a-nested-x{}

/* components/a/z.css */
.a-z{}

/* components/a-b/feature.css */
.a-b{}

/* components/ab/feature.css */
.ab{}

/* components/b/a.css */
.b-a{}
}

@layer views {

/* views/feed.css */
.feed{}

/* views/stats.css */
.stats{}
}
`, string(css))
}

func TestBuildStylesheetMissingSource(t *testing.T) {
	_, err := BuildStylesheet(fstest.MapFS{
		"styles/base.css":          {Data: []byte("")},
		"components/a/feature.css": {Data: []byte("")},
	})

	assert.Error(t, err, "views is missing")
}
