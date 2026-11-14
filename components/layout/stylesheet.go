package layout

import (
	"crypto/sha256"
	"encoding/hex"
)

// stylesheetURL is the URL the layout links the stylesheet at. It stays
// unhashed until SetStylesheet is called at startup, so tests can render the
// layout without the file on disk.
var stylesheetURL = "/public/styles.css"

// SetStylesheet points the layout at the stylesheet with a cache-busting
// version: the first 6 hex characters of the SHA-256 of its contents.
func SetStylesheet(css []byte) {
	hash := sha256.Sum256(css)
	stylesheetURL = "/public/styles.css?v=" + hex.EncodeToString(hash[:])[:6]
}
