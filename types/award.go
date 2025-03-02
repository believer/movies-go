package types

import "database/sql"

type Award struct {
	ID       string         `db:"id"`
	Detail   sql.NullString `db:"detail"`
	ImdbID   string         `db:"imdb_id"`
	Name     string         `db:"name"`
	Person   sql.NullString `db:"person"`
	PersonId sql.NullInt64  `db:"person_id"`
	Winner   bool           `db:"winner"`
	Year     string         `db:"year"`
}
