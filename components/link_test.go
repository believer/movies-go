package components

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
		_ = Link(href, "", false).Render(ctx, w)
		_ = w.Close()
	}()

	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	if actualHref, _ := doc.Find("a").Attr("href"); actualHref != href {
		t.Errorf("Expected href %q, got %q", href, actualHref)
	}

	if actualTitle := doc.Find("a").Text(); actualTitle != title {
		t.Errorf("Expected title name %q, got %q", title, actualTitle)
	}

	if doc.Find("a[_]").Length() != 0 {
		t.Errorf("Hyperscript should not be set for empty strings")
	}
}

func TestLinkWithHyperscript(t *testing.T) {
	var (
		r, w        = io.Pipe()
		href        = "/posts"
		title       = "Posts"
		hyperscript = "on click"
	)

	children := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, title)
		return err
	})

	ctx := templ.WithChildren(context.Background(), children)

	go func() {
		_ = Link(href, hyperscript, false).Render(ctx, w)
		_ = w.Close()
	}()

	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	if actualHyperscript, _ := doc.Find("a").Attr("_"); actualHyperscript != hyperscript {
		t.Errorf("Expected hyperscript %q, got %q", hyperscript, actualHyperscript)
	}
}
