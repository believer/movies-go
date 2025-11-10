package link

import (
	"context"
	"io"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ"
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
		_ = Link(Props{
			Href: href,
		}).Render(ctx, w)
		_ = w.Close()
	}()

	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	a := doc.Find("a")

	if actualHref, _ := a.Attr("href"); actualHref != href {
		t.Errorf("Expected href %q, got %q", href, actualHref)
	}

	if actualTitle := a.Text(); actualTitle != title {
		t.Errorf("Expected title name %q, got %q", title, actualTitle)
	}
}
