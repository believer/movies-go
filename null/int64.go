package null

import (
	"database/sql"
	"encoding/json"
)

type Int64 struct {
	sql.NullInt64
}

func (ni *Int64) UnmarshalJSON(b []byte) error {
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

func (ni Int64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}
