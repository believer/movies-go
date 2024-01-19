package types

import (
	"believer/movies/utils"
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
	Overview    string          `db:"overview" json:"overview"`
	Poster      string          `db:"poster"`
	Rating      sql.NullInt64   `db:"rating"`
	ReleaseDate time.Time       `db:"release_date" json:"release_date"`
	Runtime     int             `db:"runtime"`
	Tagline     string          `db:"tagline"`
	Title       string          `db:"title" json:"title"`
	UpdatedAt   time.Time       `db:"updated_at"`
	WatchedAt   time.Time       `db:"watched_at" json:"watchedAt"`
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
	return utils.FormatRuntime(m.Runtime)
}
