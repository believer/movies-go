package utils

import (
	"database/sql"
	"encoding/json"
	"time"
)

type NullTime struct {
	sql.NullTime
}

func (nt *NullTime) UnmarshalJSON(b []byte) error {
	// Handle "null"
	if string(b) == "null" {
		nt.Valid = false
		return nil
	}

	// Parse as string
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	nt.Time = t
	nt.Valid = true
	return nil
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time.Format(time.RFC3339))
}
