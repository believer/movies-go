package types

import (
	"believer/movies/utils"
	"fmt"
	"strconv"

	"github.com/a-h/templ"
)

type Stats struct {
	UniqueMovies      int `db:"unique_movies"`
	SeenWithRewatches int `db:"seen_with_rewatches"`
	TotalRuntime      int `db:"total_runtime"`
}

type MovieStats struct {
	Title string `db:"title"`
	ID    string `db:"id"`
	Count int    `db:"count"`
}

type PersonStats struct {
	Name  string `db:"name"`
	ID    string `db:"id"`
	Count int    `db:"count"`
}

type GraphData struct {
	Label string `db:"label"`
	Value int    `db:"value"`
}

type Bar struct {
	Label     string
	Value     int
	BarHeight int
	BarWidth  int
	BarX      int
	BarY      int
	LabelX    float64
	LabelY    float64
	ValueX    float64
	ValueY    int
}

type Genre struct {
	Name  string `db:"name"`
	ID    string `db:"id"`
	Count int    `db:"count"`
}

type HighestRated struct {
	Appearances           int     `db:"appearances"`
	ID                    int     `db:"id"`
	Name                  string  `db:"name"`
	TotalRating           int     `db:"total_rating"`
	WeightedAverageRating float64 `db:"weighted_average_rating"`
}

func (h HighestRated) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/person/%s-%d", utils.Slugify(h.Name), h.ID))
}

func (h HighestRated) Rank() string {
	return strconv.FormatFloat(h.WeightedAverageRating, 'f', 2, 64)
}
