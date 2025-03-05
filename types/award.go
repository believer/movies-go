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
	Name     string         `db:"name"`
	Title    sql.NullString `db:"title"`
	Person   sql.NullString `db:"person"`
	PersonId sql.NullInt64  `db:"person_id"`
	Winner   bool           `db:"winner"`
	Year     string         `db:"year"`
}

func (a *Award) YearAndName(display bool) string {
	if !display {
		return ""
	}

	if a.Title.Valid {
		return fmt.Sprintf("(%s - %s)", a.Title.String, a.Year)
	}

	return fmt.Sprintf("(%s)", a.Year)
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
