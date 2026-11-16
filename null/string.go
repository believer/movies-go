package null

import (
	"database/sql"
	"encoding/json"
)

type String struct {
	sql.NullString
}

func (ns *String) UnmarshalJSON(b []byte) error {
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

func (ns String) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}
