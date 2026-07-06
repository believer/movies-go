package utils

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func setVariables(filename string) (err error) {
	_, file, _, ok := runtime.Caller(1)

	if !ok {
		return fmt.Errorf("no calling file")
	}

	var f *os.File
	f, err = os.Open(filepath.Join(filepath.Dir(file), "..", filename))

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

	slog.Info("ENV loaded from", "File", filename)

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
