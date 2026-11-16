package awards

import "embed"

// files holds oscars.csv and baftas.csv, so the binary needs no files beside
// it to add awards.
//
//go:embed oscars.csv baftas.csv
var files embed.FS
