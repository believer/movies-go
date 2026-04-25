package link

import (
	"context"
	"io"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
)

func TestLink(t *testing.T) {
	var (
		r, w  = io.Pipe()
		href  = "/posts"
		title = "Posts"
	)

	children := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, title)
		return err
	})

	ctx := templ.WithChildren(context.Background(), children)

	go func() {
		_ = Link(Props{Href: href}).Render(ctx, w)
		_ = w.Close()
	}()

	doc, err := goquery.NewDocumentFromReader(r)
	assert.NoError(t, err, "Failed to read template")

	a := doc.Find("a")
	actualHref, _ := a.Attr("href")
	assert.Equal(t, href, actualHref)
	assert.Equal(t, title, a.Text())
}
