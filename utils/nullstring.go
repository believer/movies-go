package utils

import (
	"database/sql"
	"encoding/json"
)

type NullString struct {
	sql.NullString
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	var s *string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s != nil {
		ns.String = *s
		ns.Valid = true
	} else {
		ns.String = ""
		ns.Valid = false
	}
	return nil
}
