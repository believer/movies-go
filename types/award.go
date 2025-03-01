package types

import "database/sql"

type Award struct {
	ID     string         `db:"id"`
	ImdbID string         `db:"imdb_id"`
	Name   string         `db:"name"`
	Person sql.NullString `db:"person"`
	Winner bool           `db:"winner"`
	Year   string         `db:"year"`
}
