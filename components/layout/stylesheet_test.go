package layout

import (
	"context"
	"io"
	"testing"

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

	assert.Equal(t, "/public/styles.css", renderStylesheetHref(t))

	// SHA-256 of "body{}" starts with 7c9804, of "body{color:red}" with 15c42a
	VersionStylesheet([]byte("body{}"))
	assert.Equal(t, "/public/styles.css?v=7c9804", renderStylesheetHref(t))

	VersionStylesheet([]byte("body{color:red}"))
	assert.Equal(t, "/public/styles.css?v=15c42a", renderStylesheetHref(t))
}
