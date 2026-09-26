package tests

import (
	"bytes"
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
	"golang.org/x/net/html"
)

func outerHTML(s *goquery.Selection) (string, error) {
	if s.Length() == 0 {
		return "", fmt.Errorf("empty selection")
	}
	var buf bytes.Buffer
	err := html.Render(&buf, s.Get(0))
	return buf.String(), err
}

func DebugHtml(s *goquery.Selection) {
	if out, err := outerHTML(s); err == nil {
		log.Println(gohtml.Format(out))
	}
}
