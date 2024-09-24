package types

import (
	"database/sql"
	"fmt"

	"github.com/a-h/templ"
)

type Series struct {
	ID           int           `db:"id"`
	Name         string        `db:"name"`
	ParentSeries sql.NullInt64 `db:"parent_series"`
}

// Link to series
func (s Series) LinkToParent() templ.SafeURL {
	if s.ParentSeries.Valid {
		return templ.URL(fmt.Sprintf("/series/%d", s.ParentSeries.Int64))
	}

	return ""
}

type SeriesMovies struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Movies Movies `db:"movies"`
}

// Link to series
func (s SeriesMovies) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/series/%d", s.ID))
}
