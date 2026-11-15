package intake

import (
	"log/slog"
	"os"
	"testing"
)

// Intake logs every step of adding a movie. Keep test output to warnings and
// failures.
func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelWarn)
	os.Exit(m.Run())
}
