package button

import (
	"bytes"
	"context"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ"
)

func setup(props Props) (*goquery.Document, error) {
	var buf bytes.Buffer

	// Add child
	child := templ.Raw("Save changes")
	ctx := templ.WithChildren(context.Background(), child)

	_ = Button(props).Render(ctx, &buf)

	return goquery.NewDocumentFromReader(&buf)
}

func TestButton_Primary(t *testing.T) {
	doc, err := setup(Props{})
	if err != nil {
		t.Fatal(err)
	}

	btn := doc.Find("button[type='button']")

	hasFocusClass := btn.HasClass("focus")
	hasBtnClass := btn.HasClass("button")
	isSecondary := btn.HasClass("secondary")

	if !hasBtnClass || !hasFocusClass {
		t.Error("missing default classes")
	}

	if isSecondary {
		t.Error("button is not primary")
	}

	// Only render one button
	if btn.Length() != 1 {
		t.Fatalf("expected exactly one button, got %d", btn.Length())
	}

	// Loading text is alwayas present, but hidden
	if btn.Find(".loading").Text() != "Sending..." {
		t.Error("expected loading text")
	}
}

func TestButton_Secondary(t *testing.T) {
	doc, err := setup(Props{Rank: Secondary})
	if err != nil {
		t.Fatal(err)
	}

	btn := doc.Find("button[type='button']")

	if !btn.HasClass("secondary") {
		t.Error("expected secondary class")
	}
}

func TestButton_Submit(t *testing.T) {
	doc, err := setup(Props{Type: TypeSubmit})
	if err != nil {
		t.Fatal(err)
	}

	btn := doc.Find("button[type='submit']")

	if btn.Length() != 1 {
		t.Fatalf("expected exactly one button, got %d", btn.Length())
	}
}

func TestButton_CustomAttributes(t *testing.T) {
	doc, err := setup(Props{Attrs: templ.Attributes{
		"data-testid": "hello",
	}})
	if err != nil {
		t.Fatal(err)
	}

	btn := doc.Find("button[type='button']")

	if val, ok := btn.Attr("data-testid"); !ok || val != "hello" {
		t.Fatalf("expected exactly one button, got %d", btn.Length())
	}
}
