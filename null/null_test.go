package null

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringJSON(t *testing.T) {
	tests := []struct {
		json string
		want String
	}{
		{`"Nolan"`, String{sql.NullString{String: "Nolan", Valid: true}}},
		{`null`, String{}},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var got String
			require.NoError(t, json.Unmarshal([]byte(tc.json), &got))
			assert.Equal(t, tc.want, got)

			b, err := json.Marshal(got)
			require.NoError(t, err)
			assert.JSONEq(t, tc.json, string(b))
		})
	}
}

func TestInt64JSON(t *testing.T) {
	tests := []struct {
		json string
		want Int64
	}{
		{`4`, Int64{sql.NullInt64{Int64: 4, Valid: true}}},
		{`null`, Int64{}},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var got Int64
			require.NoError(t, json.Unmarshal([]byte(tc.json), &got))
			assert.Equal(t, tc.want, got)

			b, err := json.Marshal(got)
			require.NoError(t, err)
			assert.JSONEq(t, tc.json, string(b))
		})
	}
}
