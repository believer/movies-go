package types

import (
	"believer/movies/format"
	"believer/movies/slug"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Stats struct {
	UniqueMovies      int `db:"unique_movies"`
	SeenWithRewatches int `db:"seen_with_rewatches"`
	TotalRuntime      int `db:"total_runtime"`
}

func (s *Stats) FormattedNumberOfMoviesSeen() string {
	return format.Number(s.UniqueMovies)
}

func (s *Stats) FormattedNumberOfMoviesSeenWithRewatches() string {
	return format.Number(s.SeenWithRewatches)
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

func (h HighestRated) LinkTo() string {
	return fmt.Sprintf("/person/%s-%d", slug.Make(h.Name), h.ID)
}

func (h HighestRated) Rank() string {
	return strconv.FormatFloat(h.WeightedAverageRating, 'f', 2, 64)
}

// HotTake is a movie a user rated away from its IMDb rating.
type HotTake struct {
	ID         int     `db:"id"`
	Title      string  `db:"title"`
	Rating     int     `db:"rating"`
	IMDbRating float64 `db:"imdb_rating"`
	// Difference is the user's rating minus IMDb's
	Difference float64 `db:"difference"`
}

func (h HotTake) LinkTo() string {
	return fmt.Sprintf("/movie/%s-%d", slug.Make(h.Title), h.ID)
}

// FormattedDifference is the difference with its sign, like +3.1 or −6.3.
func (h HotTake) FormattedDifference() string {
	return strings.Replace(fmt.Sprintf("%+.1f", h.Difference), "-", "−", 1)
}

// IMDbComparison is how a user's ratings compare with IMDb's, across the
// movies that have both.
type IMDbComparison struct {
	Movies            int     `db:"movies"`
	AverageDifference float64 `db:"average_difference"`
}

// Summary says which side of IMDb the user rates on, and by how much.
func (c IMDbComparison) Summary() string {
	rounded := fmt.Sprintf("%.1f", math.Abs(c.AverageDifference))

	switch {
	case rounded == "0.0":
		return "Level with IMDb on average"
	case c.AverageDifference > 0:
		return rounded + " above IMDb on average"
	default:
		return rounded + " below IMDb on average"
	}
}

// Streaks are records of a user's watching days. A record a user has no
// Watches for has 0 days or watches and no dates.
type Streaks struct {
	StreakDays        int          `db:"streak_days"`
	StreakStart       sql.NullTime `db:"streak_start"`
	StreakEnd         sql.NullTime `db:"streak_end"`
	BreakDays         int          `db:"break_days"`
	BreakStart        sql.NullTime `db:"break_start"`
	BreakEnd          sql.NullTime `db:"break_end"`
	BusiestDayWatches int          `db:"busiest_day_watches"`
	BusiestDay        sql.NullTime `db:"busiest_day"`
}

// PersonPair is an actor and a director who made movies together.
type PersonPair struct {
	ActorID      int    `db:"actor_id"`
	ActorName    string `db:"actor_name"`
	DirectorID   int    `db:"director_id"`
	DirectorName string `db:"director_name"`
	Movies       int    `db:"movies"`
}

func (p PersonPair) ActorLink() string {
	return fmt.Sprintf("/person/%s-%d", slug.Make(p.ActorName), p.ActorID)
}

func (p PersonPair) DirectorLink() string {
	return fmt.Sprintf("/person/%s-%d", slug.Make(p.DirectorName), p.DirectorID)
}

type OthersStats struct {
	Seen          int     `db:"seen_count" json:"seenByUsers"`
	AverageRating float64 `db:"avg_rating" json:"averageRating"`
}
