package utils

import (
	"database/sql"
	"encoding/json"
)

type NullInt64 struct {
	sql.NullInt64
}

func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	var i *int64
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	if i != nil {
		ni.Int64 = *i
		ni.Valid = true
	} else {
		ni.Int64 = 0
		ni.Valid = false
	}
	return nil
}
