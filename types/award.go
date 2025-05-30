package types

import (
	"believer/movies/utils"
	"fmt"

	"github.com/a-h/templ"
)

type Award struct {
	Category string           `db:"category" json:"category"`
	Detail   utils.NullString `db:"detail" json:"detail"`
	ID       string           `db:"id" json:"id"`
	ImdbID   string           `db:"imdb_id"`
	MovieID  int              `db:"movie_id"`
	Nominees Nominees         `db:"nominees"`
	Person   utils.NullString `db:"person" json:"person"`
	PersonId utils.NullInt64  `db:"person_id" json:"person_id"`
	Title    utils.NullString `db:"title"`
	Winner   bool             `db:"winner" json:"winner"`
	Year     string           `db:"year"`
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

func (a *Award) LinkToPerson() templ.SafeURL {
	if a.Person.Valid && a.PersonId.Valid {
		return templ.SafeURL(fmt.Sprintf("/person/%s-%d", utils.Slugify(a.Person.String), a.PersonId.Int64))
	}

	return "#"
}

func (a *Award) LinkToYear() templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("/awards/year/%s", a.Year))
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

type GlobalAward struct {
	MovieID int          `db:"movie_id"`
	Title   string       `db:"title"`
	Awards  GlobalAwards `db:"awards"`
}

type GlobalAwards []Award

func (u *GlobalAwards) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

func (g *GlobalAward) LinkToMovie() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/movie/%s-%d", utils.Slugify(g.Title), g.MovieID))
}
