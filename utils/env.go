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
func setVariables(filename string) (err error) {
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

		if k, v, ok := strings.Cut(line, "="); ok {
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

func LoadEnv() {
	if err := setVariables(".env"); err != nil {
		slog.Warn("Could not load .env file", "error", err)
	}
	if err := setVariables(".env.local"); err != nil {
		slog.Warn("Could not load .env.local file", "error", err)
	}
}
