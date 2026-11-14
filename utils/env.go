package utils

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// setVariables loads the file from the working directory, so a binary
// built in one checkout and run from another reads the second one's env
func setVariables(filename string, fromProcess map[string]bool) (err error) {
	path, err := filepath.Abs(filename)

	if err != nil {
		return err
	}

	var f *os.File
	f, err = os.Open(path)

	if err != nil {
		return err
	}

	defer func() {
		cerr := f.Close()

		if err == nil {
			err = cerr
		}
	}()

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if k, v, ok := strings.Cut(line, "="); ok && !fromProcess[k] {
			err := os.Setenv(k, strings.Trim(v, `"`))

			if err != nil {
				return err
			}
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	slog.Info("ENV loaded from", "File", path)

	return nil
}

// LoadEnv loads .env and then .env.local from the working directory. A
// non-empty variable from the process wins over both, so a server started
// with its own DATABASE_URL, like the E2E server, keeps it.
func LoadEnv() {
	fromProcess := map[string]bool{}

	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok && v != "" {
			fromProcess[k] = true
		}
	}

	if err := setVariables(".env", fromProcess); err != nil {
		slog.Warn("Could not load .env file", "error", err)
	}
	if err := setVariables(".env.local", fromProcess); err != nil {
		slog.Warn("Could not load .env.local file", "error", err)
	}
}
