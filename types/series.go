package types

import (
	"believer/movies/utils"
	"encoding/json"
	"fmt"

	"github.com/a-h/templ"
)

type ParentSeries struct {
	Name string `db:"name"`
	ID   int    `db:"int"`
}

type ParentSeriesMult []ParentSeries

func (u *ParentSeriesMult) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

type Series struct {
	ID           int              `db:"id"`
	Name         string           `db:"name"`
	ParentSeries ParentSeriesMult `db:"parent_series"`
}

// Link to series
func (s Series) LinkToParent(id int) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/series/%s-%d", utils.Slugify(s.Name), id))
}

type SeriesMovies struct {
	ID     int            `db:"id"`
	Name   string         `db:"name"`
	Movies MoviesInSeries `db:"movies"`
}

type MoviesInSeries Movies

func (u *MoviesInSeries) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

// Link to series
func (s SeriesMovies) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/series/%s-%d", utils.Slugify(s.Name), s.ID))
}
