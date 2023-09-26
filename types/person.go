package types

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Cast     Movies `json:"cast" db:"cast"`
	Director Movies `json:"director" db:"director"`
	Writer   Movies `json:"writer" db:"writer"`
	Composer Movies `json:"composer" db:"composer"`
	Producer Movies `json:"producer" db:"producer"`
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
