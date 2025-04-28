package types

import (
	"believer/movies/utils"
	"fmt"

	"github.com/a-h/templ"
)

type ParentSeries struct {
	Name string `db:"name"`
	ID   int    `db:"int"`
}

type ParentSeriesMult []ParentSeries

func (u *ParentSeriesMult) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

func seriesLink(name string, id int) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/series/%s-%d", utils.Slugify(name), id))
}

func (p *ParentSeries) LinkTo() templ.SafeURL {
	return seriesLink(p.Name, p.ID)
}

type Series struct {
	ID           int              `db:"id"`
	Name         string           `db:"name"`
	ParentSeries ParentSeriesMult `db:"parent_series"`
}

// Link to series
func (s Series) LinkToParent(id int) templ.SafeURL {
	return seriesLink(s.Name, id)
}

type SeriesMovies struct {
	ID     int            `db:"id"`
	Name   string         `db:"name"`
	Movies MoviesInSeries `db:"movies"`
}

type MoviesInSeries Movies

func (u *MoviesInSeries) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

// Link to series
func (s SeriesMovies) LinkTo() templ.SafeURL {
	return seriesLink(s.Name, s.ID)
}
