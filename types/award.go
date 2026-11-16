package types

import (
	"believer/movies/null"
	"believer/movies/slug"
	"fmt"
)

// Award sub structs
// ======================================================

type Nominees []Person

func (u *Nominees) Scan(v any) error {
	return scanJSON(v, u)
}

// Award
// ======================================================

type Award struct {
	Category string      `db:"category" json:"category"`
	Detail   null.String `db:"detail" json:"detail"`
	ID       string      `db:"id" json:"id"`
	ImdbID   string      `db:"imdb_id" json:"imdbId"`
	MovieID  null.Int64  `db:"movie_id" json:"movieId"`
	Nominees Nominees    `db:"nominees" json:"nominees"`
	Person   null.String `db:"person" json:"person"`
	PersonId null.Int64  `db:"person_id" json:"personId"`
	Title    null.String `db:"title" json:"title"`
	Type     string      `db:"type" json:"type"`
	Winner   bool        `db:"winner" json:"winner"`
	Year     string      `db:"year" json:"year"`
}

type Awards []Award

func (u *Awards) Scan(v any) error {
	return scanJSON(v, u)
}

func (a *Award) LinkToMovie() string {
	if a.Title.Valid && a.MovieID.Valid {
		return fmt.Sprintf("/movie/%s-%d", slug.Make(a.Title.String), a.MovieID.Int64)
	}

	return "#"
}

func (a *Award) LinkToPerson() string {
	if a.Person.Valid && a.PersonId.Valid {
		return fmt.Sprintf("/person/%s-%d", slug.Make(a.Person.String), a.PersonId.Int64)
	}

	return "#"
}

func (a *Award) LinkToYear() string {
	return fmt.Sprintf("/awards/year/%s", a.Year)
}

// Awards for person
// ======================================================

type AwardPersonStat struct {
	Count int    `db:"count" json:"count"`
	ID    int    `db:"person_id" json:"id"`
	Name  string `db:"person" json:"name"`
}

func (a AwardPersonStat) LinkTo() string {
	return fmt.Sprintf("/person/%s-%d", slug.Make(a.Name), a.ID)
}

// Awards for movie
// ======================================================

type AwardMovieStat struct {
	Count int    `db:"award_count" json:"count"`
	ID    int    `db:"id" json:"id"`
	Title string `db:"title" json:"title"`
}

func (a AwardMovieStat) LinkTo() string {
	return fmt.Sprintf("/movie/%s-%d", slug.Make(a.Title), a.ID)
}

// Awards for person
// ======================================================

type GroupedAward struct {
	Name     string `json:"name"`
	Winner   bool   `json:"winner"`
	Nominees Awards `json:"nominees"`
}

type GroupedAwards map[string]GroupedAward

// Awards
// ======================================================

type AwardsByYear struct {
	MovieID int    `db:"movie_id" json:"movieId"`
	Title   string `db:"title" json:"title"`
	Awards  Awards `db:"awards" json:"awards"`
}

func (g *AwardsByYear) LinkToMovie() string {
	return fmt.Sprintf("/movie/%s-%d", slug.Make(g.Title), g.MovieID)
}

// Awards by category
// ======================================================

type AwardsByCategory struct {
	Category string `db:"category" json:"category"`
	Nominees Awards `db:"nominees" json:"nominees"`
}

// Award texts
// ======================================================

type AwardConfig struct {
	WinName        string
	NominationName string
	EmptyState     string
}

var awardConfigs = map[string]AwardConfig{
	"academy-award": {
		WinName:        "%d Academy Award wins",
		NominationName: "%d Academy Award nominations",
		EmptyState:     "No movies with this amount of Academy Awards",
	},
	"bafta": {
		WinName:        "%d BAFTA wins",
		NominationName: "%d BAFTA nominations",
		EmptyState:     "No movies with this amount of BAFTAs",
	},
}

func GetAwardConfig(awardType string) (cfg AwardConfig, ok bool) {
	cfg, ok = awardConfigs[awardType]
	return
}
