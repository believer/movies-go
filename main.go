package main

import (
	"believer/movies/app"
	"embed"
)

// css holds the stylesheet sources. go:embed can't reach outside the
// package's directory, so they are embedded here at the root. Whole
// directories are embedded, as patterns can't match CSS files at any
// depth, and the stylesheet build keeps only the CSS files.
//
//go:embed styles components views
var css embed.FS

func main() {
	err := app.SetupAndRunApp(css)

	if err != nil {
		panic(err)
	}
}
