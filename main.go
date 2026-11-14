package main

import (
	"believer/movies/app"
	"embed"
)

// css holds the stylesheet sources. go:embed can't reach outside the
// package's directory, so they are embedded here at the root.
//
//go:embed styles/*.css components/*/*.css views/*.css
var css embed.FS

func main() {
	err := app.SetupAndRunApp(css)

	if err != nil {
		panic(err)
	}
}
