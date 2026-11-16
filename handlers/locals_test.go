package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Handlers read the login state through Request, in request.go. A direct read can get
// a missing value wrong: Locals("IsAuthenticated") == false lets nil through,
// and Locals("UserId").(string) panics on it.
func TestHandlersReadLoginStateThroughRequest(t *testing.T) {
	files, err := filepath.Glob("*.go")
	assert.NoError(t, err)

	for _, file := range files {
		// Request itself reads them
		if strings.HasSuffix(file, "_test.go") || file == "request.go" {
			continue
		}

		src, err := os.ReadFile(file)
		assert.NoError(t, err)

		for _, local := range []string{`Locals("IsAuthenticated"`, `Locals("UserId"`} {
			if strings.Contains(string(src), local) {
				t.Errorf("%s reads %s), use NewRequest(c) instead", file, local)
			}
		}
	}
}
