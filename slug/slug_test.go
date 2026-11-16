package slug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMake(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"Hugh Jackman", "hugh-jackman"},
		{"Alfonso Cuarón", "alfonso-cuaron"},
	}
	for _, tc := range tests {
		t.Run(tc.text, func(t *testing.T) {
			assert.Equal(t, tc.want, Make(tc.text))
		})
	}
}

func TestURL(t *testing.T) {
	got := URL("root", "name", "1234")
	want := "/root/name-1234"

	assert.Equal(t, want, got)
}

func TestID(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/root/sub/test-1234", "1234"},
		{"/root/sub/test-test-250-test-1234", "1234"},
		{"/root/sub/test-test", "test"},
		{"/root/sub/united-states-of-america-US", "US"},
		{"/root/sub/1234", "1234"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.want, ID(tc.path))
		})
	}
}

func TestNumericID(t *testing.T) {
	got, err := NumericID("/root/sub/test-1234")

	assert.NoError(t, err)
	assert.Equal(t, "1234", got)

	_, err = NumericID("/root/sub/test-test")

	assert.EqualError(t, err, "not a valid ID")
}
