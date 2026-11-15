package utils

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNullStringJSON(t *testing.T) {
	tests := []struct {
		json string
		want NullString
	}{
		{`"Nolan"`, NullString{sql.NullString{String: "Nolan", Valid: true}}},
		{`null`, NullString{}},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var got NullString
			require.NoError(t, json.Unmarshal([]byte(tc.json), &got))
			assert.Equal(t, tc.want, got)

			b, err := json.Marshal(got)
			require.NoError(t, err)
			assert.JSONEq(t, tc.json, string(b))
		})
	}
}

func TestNullInt64JSON(t *testing.T) {
	tests := []struct {
		json string
		want NullInt64
	}{
		{`4`, NullInt64{sql.NullInt64{Int64: 4, Valid: true}}},
		{`null`, NullInt64{}},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var got NullInt64
			require.NoError(t, json.Unmarshal([]byte(tc.json), &got))
			assert.Equal(t, tc.want, got)

			b, err := json.Marshal(got)
			require.NoError(t, err)
			assert.JSONEq(t, tc.json, string(b))
		})
	}
}
