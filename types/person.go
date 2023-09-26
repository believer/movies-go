package types

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Id     int    `json:"id" db:"id"`
	Name   string `json:"name" db:"name"`
	Movies Movies `json:"movies" db:"movies"`
}

type Persons []Person

func (u *Persons) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

type Cast struct {
	Job    string  `db:"job"`
	Person Persons `db:"person"`
}
