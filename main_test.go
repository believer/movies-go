package main

import (
	"believer/movies/components/layout"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Every CSS file on disk ends up in the stylesheet, so a narrower embed
// pattern can't silently drop one
func TestEmbeddedStylesheet(t *testing.T) {
	stylesheet, err := layout.BuildStylesheet(css)

	assert.NoError(t, err)

	for _, dir := range []string{"styles", "components", "views"} {
		err := fs.WalkDir(os.DirFS("."), dir, func(name string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || path.Ext(name) != ".css" {
				return err
			}

			assert.Contains(t, string(stylesheet), "/* "+name+" */")

			return nil
		})

		assert.NoError(t, err)
	}
}
