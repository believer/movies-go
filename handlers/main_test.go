package handlers

import (
	"log/slog"
	"os"
	"testing"
)

// Movie intake logs every step of adding a movie. Keep test output to
// warnings and failures.
func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelWarn)
	os.Exit(m.Run())
}
