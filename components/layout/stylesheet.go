package layout

import (
	"crypto/sha256"
	"encoding/hex"
)

const stylesheetPath = "/public/styles.css"

// stylesheetURL is the URL the layout links the stylesheet at. It stays
// unversioned until VersionStylesheet is called at startup, so tests can
// render the layout without the file on disk.
var stylesheetURL = stylesheetPath

// VersionStylesheet adds a cache-busting version to the stylesheet URL: the
// first 6 hex characters of the SHA-256 of its contents.
func VersionStylesheet(css []byte) {
	hash := sha256.Sum256(css)
	stylesheetURL = stylesheetPath + "?v=" + hex.EncodeToString(hash[:])[:6]
}
