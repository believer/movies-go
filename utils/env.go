package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2/log"
)

func LoadEnv() error {
	_, file, _, ok := runtime.Caller(1)

	if !ok {
		return fmt.Errorf("No calling file")
	}

	f, err := os.Open(filepath.Join(filepath.Dir(file), "..", ".env"))

	if err != nil {
		return err
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if k, v, ok := strings.Cut(sc.Text(), "="); ok {
			os.Setenv(k, strings.Trim(v, `"`))
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	log.Info("ENV loaded from .env")

	return nil
}
