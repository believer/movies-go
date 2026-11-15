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

// A non-empty variable set before the app starts wins over both env files, so the
// E2E server's DATABASE_URL isn't replaced by a checkout's .env.local
func TestLoadEnvKeepsProcessEnv(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("ENV_TEST_PROCESS", "process")
	t.Setenv("ENV_TEST_FILE", "")

	write := func(name, content string) {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
		assert.NoError(t, err)
	}

	write(".env", "ENV_TEST_PROCESS=env\nENV_TEST_FILE=env\n")
	write(".env.local", "ENV_TEST_PROCESS=local\nENV_TEST_FILE=local\n")

	LoadEnv()

	assert.Equal(t, "process", os.Getenv("ENV_TEST_PROCESS"))
	assert.Equal(t, "local", os.Getenv("ENV_TEST_FILE"))
}
