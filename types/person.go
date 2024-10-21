package types

import (
	"believer/movies/utils"
	"encoding/json"
	"fmt"
	"time"

	"github.com/a-h/templ"
)

type PersonMovie struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	ReleaseDate time.Time `json:"release_date" db:"release_date"`
	Seen        bool      `json:"seen" db:"seen"`
	Character   string    `json:"character" db:"character"`
}

// Link to the movie
func (m PersonMovie) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/movie/%d", m.ID))
}

// Release year
func (m PersonMovie) ReleaseYear() string {
	return m.ReleaseDate.Format("2006")
}

// The movie's release date formatted as ISO 8601 - YYYY-MM-DD
func (m PersonMovie) ISOReleaseDate() string {
	return m.ReleaseDate.Format("2006-01-02")
}

// Link to the movie's release year
func (m PersonMovie) LinkToYear() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/year/%s", m.ReleaseDate.Format("2006")))
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

// Link to the person
func (p Person) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/person/%s-%d", utils.Slugify(p.Name), p.ID))
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
