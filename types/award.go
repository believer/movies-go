package types

import (
	"believer/movies/utils"
	"fmt"
)

// Award sub structs
// ======================================================

type Nominees []Person

func (u *Nominees) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

// Award
// ======================================================

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

type Awards []Award

func (u *Awards) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

func (a *Award) LinkToMovie() string {
	if a.Title.Valid {
		return fmt.Sprintf("/movie/%s-%d", utils.Slugify(a.Title.String), a.MovieID)
	}

	return "#"
}

func (a *Award) LinkToPerson() string {
	if a.Person.Valid && a.PersonId.Valid {
		return fmt.Sprintf("/person/%s-%d", utils.Slugify(a.Person.String), a.PersonId.Int64)
	}

	return "#"
}

func (a *Award) LinkToYear() string {
	return fmt.Sprintf("/awards/year/%s", a.Year)
}

// Awards for person
// ======================================================

type AwardPersonStat struct {
	Count int    `db:"count"`
	ID    int    `db:"person_id"`
	Name  string `db:"person"`
}

func (a AwardPersonStat) LinkTo() string {
	return fmt.Sprintf("/person/%s-%d", utils.Slugify(a.Name), a.ID)
}

// Awards for movie
// ======================================================

type AwardMovieStat struct {
	Count int    `db:"award_count"`
	ID    int    `db:"id"`
	Title string `db:"title"`
}

func (a AwardMovieStat) LinkTo() string {
	return fmt.Sprintf("/movie/%s-%d", utils.Slugify(a.Title), a.ID)
}

// Awards for person
// ======================================================

type GroupedAward struct {
	Name     string
	Winner   bool
	Nominees []Award
}

type GroupedAwards map[string]GroupedAward

// Awards
// ======================================================

type AwardsByYear struct {
	MovieID int    `db:"movie_id"`
	Title   string `db:"title"`
	Awards  Awards `db:"awards"`
}

func (g *AwardsByYear) LinkToMovie() string {
	return fmt.Sprintf("/movie/%s-%d", utils.Slugify(g.Title), g.MovieID)
}
