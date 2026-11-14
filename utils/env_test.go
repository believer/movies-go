package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The env files come from the working directory, not from the checkout
// the binary was built in, and .env.local wins over .env
func TestLoadEnv(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("ENV_TEST_SHARED", "")
	t.Setenv("ENV_TEST_ONLY_ENV", "")

	write := func(name, content string) {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
		assert.NoError(t, err)
	}

	write(".env", "ENV_TEST_SHARED=env\nENV_TEST_ONLY_ENV=env\n")
	write(".env.local", "ENV_TEST_SHARED=\"local\"\n")

	LoadEnv()

	assert.Equal(t, "local", os.Getenv("ENV_TEST_SHARED"))
	assert.Equal(t, "env", os.Getenv("ENV_TEST_ONLY_ENV"))
}
