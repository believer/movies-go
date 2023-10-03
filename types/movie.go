package types

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type CastAndCrew struct {
	Name string `db:"name"`
	Job  string `db:"job"`
}

type Movie struct {
	Cast        []CastAndCrew   `db:"cast"`
	CreatedAt   time.Time       `db:"created_at"`
	Genres      pq.StringArray  `db:"genres"`
	ID          int             `db:"id" json:"id"`
	ImdbId      string          `db:"imdb_id"`
	ImdbRating  sql.NullFloat64 `db:"imdb_rating"`
	Overview    string          `db:"overview"`
	Poster      string          `db:"poster"`
	Rating      int             `db:"rating"`
	ReleaseDate time.Time       `db:"release_date" json:"release_date"`
	Runtime     int             `db:"runtime"`
	Tagline     string          `db:"tagline"`
	Title       string          `db:"title" json:"title"`
	UpdatedAt   time.Time       `db:"updated_at"`
	WatchedAt   time.Time       `db:"watched_at"`
}

type Movies []Movie

func (u *Movies) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

// Format runtime in hours and minutes from minutes
func (m Movie) RuntimeFormatted() string {
	hours := m.Runtime / 60
	minutes := m.Runtime % 60

	return fmt.Sprintf("%dh %dm", hours, minutes)
}
