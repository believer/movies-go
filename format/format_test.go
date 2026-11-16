package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuntime(t *testing.T) {
	tests := []struct {
		runtime int
		want    string
	}{
		{0, "0m"},
		{1, "1m"},
		{2, "2m"},
		{60, "1h"},
		{61, "1h 1m"},
		{120, "2h"},
		{121, "2h 1m"},
		{1440, "1d"},
		{1441, "1d 1m"},
		{1500, "1d 1h"},
		{1501, "1d 1h 1m"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, Runtime(tc.runtime))
		})
	}
}

func TestNumber(t *testing.T) {
	assert.Equal(t, "0", Number(0))
	assert.Equal(t, "999", Number(999))
	assert.Equal(t, "12,345", Number(12345))
	assert.Equal(t, "1,234,567", Number(1234567))
}
