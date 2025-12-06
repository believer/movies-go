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

func LoadEnv() error {
	_, file, _, ok := runtime.Caller(1)

	if !ok {
		return fmt.Errorf("no calling file")
	}

	f, err := os.Open(filepath.Join(filepath.Dir(file), "..", ".env"))

	if err != nil {
		return err
	}

	defer func() {
		cerr := f.Close()

		if err != nil {
			err = cerr
		}
	}()

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if k, v, ok := strings.Cut(sc.Text(), "="); ok {
			err := os.Setenv(k, strings.Trim(v, `"`))

			if err != nil {
				return err
			}
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	slog.Info("ENV loaded from .env")

	return nil
}
