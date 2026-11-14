package layout

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
)

const stylesheetPath = "/styles.css"

// stylesheetURL is the URL the layout links the stylesheet at. It stays
// unversioned until VersionStylesheet is called at startup, so tests can
// render the layout without the stylesheet built.
var stylesheetURL = stylesheetPath

// VersionStylesheet adds a cache-busting version to the stylesheet URL: the
// first 6 hex characters of the SHA-256 of its contents.
func VersionStylesheet(css []byte) {
	hash := sha256.Sum256(css)
	stylesheetURL = stylesheetPath + "?v=" + hex.EncodeToString(hash[:])[:6]
}

// stylesheetSources are the directories the stylesheet is built from, in
// order. Files in a directory with a layer are wrapped in it, the others
// declare their own layers.
var stylesheetSources = []struct {
	dir   string
	layer string
}{
	{dir: "styles"},
	{dir: "components", layer: "components"},
	{dir: "views", layer: "views"},
}

// BuildStylesheet joins the CSS files in fsys into one stylesheet. It
// declares the layer order first, then adds each source directory's files
// in the order fs.WalkDir visits them.
func BuildStylesheet(fsys fs.FS) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString("@layer base, components, views;\n")

	for _, source := range stylesheetSources {
		if source.layer != "" {
			fmt.Fprintf(&b, "\n@layer %s {\n", source.layer)
		}

		err := fs.WalkDir(fsys, source.dir, func(name string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || path.Ext(name) != ".css" {
				return err
			}

			css, err := fs.ReadFile(fsys, name)

			if err != nil {
				return err
			}

			fmt.Fprintf(&b, "\n/* %s */\n", name)
			b.Write(css)

			return nil
		})

		if err != nil {
			return nil, err
		}

		if source.layer != "" {
			b.WriteString("}\n")
		}
	}

	return b.Bytes(), nil
}
