package types

import (
	"encoding/json"
	"fmt"
	"time"
)

type PersonMovie struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	ReleaseDate time.Time `json:"release_date" db:"release_date"`
	Seen        bool      `json:"seen" db:"seen"`
	Character   string    `json:"character" db:"character"`
}

type PersonMovies []PersonMovie

func (u *PersonMovies) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

type Person struct {
	ID       int          `json:"id" db:"id"`
	Name     string       `json:"name" db:"name"`
	Cast     PersonMovies `json:"cast" db:"cast"`
	Director PersonMovies `json:"director" db:"director"`
	Writer   PersonMovies `json:"writer" db:"writer"`
	Composer PersonMovies `json:"composer" db:"composer"`
	Producer PersonMovies `json:"producer" db:"producer"`
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
