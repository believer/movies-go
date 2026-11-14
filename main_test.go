package main

import (
	"believer/movies/components/layout"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddedStylesheet(t *testing.T) {
	stylesheet, err := layout.BuildStylesheet(css)

	assert.NoError(t, err)
	assert.Contains(t, string(stylesheet), "/* components/button/button.css */")
	assert.Contains(t, string(stylesheet), "/* views/feed.css */")
}
