package types

import (
	"database/sql"
	"fmt"
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
