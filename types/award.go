package types

import (
	"believer/movies/utils"
	"database/sql"
	"fmt"

	"github.com/a-h/templ"
)

type Award struct {
	ID       string         `db:"id"`
	Detail   sql.NullString `db:"detail"`
	ImdbID   string         `db:"imdb_id"`
	MovieID  int            `db:"movie_id"`
	Category string         `db:"category"`
	Nominees Nominees       `db:"nominees"`
	Title    sql.NullString `db:"title"`
	Person   sql.NullString `db:"person"`
	PersonId sql.NullInt64  `db:"person_id"`
	Winner   bool           `db:"winner"`
	Year     string         `db:"year"`
}

type Nominees []Person

func (u *Nominees) Scan(v any) error {
	return ScanEntity(v, u)
}

func (a *Award) LinkToMovie() templ.SafeURL {
	if a.Title.Valid {
		return templ.SafeURL(fmt.Sprintf("/movie/%s-%d", utils.Slugify(a.Title.String), a.MovieID))
	}

	return "#"
}

func (a *Award) LinkToYear() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("/year/%s", a.Year))
}

type AwardPersonStat struct {
	Count int    `db:"count"`
	ID    int    `db:"person_id"`
	Name  string `db:"person"`
}

func (a AwardPersonStat) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/person/%s-%d", utils.Slugify(a.Name), a.ID))
}

type AwardMovieStat struct {
	Count int    `db:"award_count"`
	ID    int    `db:"id"`
	Title string `db:"title"`
}

func (a AwardMovieStat) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/movie/%s-%d", utils.Slugify(a.Title), a.ID))
}

type GroupedAward struct {
	Name     string
	Winner   bool
	Nominees []Award
}

type GroupedAwards map[string]GroupedAward
