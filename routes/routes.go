package routes

import (
	"believer/movies/db"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type CastAndCrew struct {
	Name string `db:"name"`
	Job  string `db:"job"`
}

type Movie struct {
	Cast        []CastAndCrew   `db:"cast"`
	CreatedAt   time.Time       `db:"created_at"`
	Genres      pq.StringArray  `db:"genres"`
	Id          int             `db:"id"`
	ImdbId      string          `db:"imdb_id"`
	ImdbRating  sql.NullFloat64 `db:"imdb_rating"`
	Overview    string          `db:"overview"`
	Poster      string          `db:"poster"`
	ReleaseDate time.Time       `db:"release_date"`
	Runtime     int             `db:"runtime"`
	Tagline     string          `db:"tagline"`
	Title       string          `db:"title"`
	UpdatedAt   time.Time       `db:"updated_at"`
	WatchedAt   time.Time       `db:"watched_at"`
}

// Format runtime in hours and minutes from minutes
func (m Movie) RuntimeFormatted() string {
	hours := m.Runtime / 60
	minutes := m.Runtime % 60

	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func FeedHandler(c *fiber.Ctx) error {
	var movies []Movie

	err := db.Client.Select(&movies, `
SELECT m.id, m.title, m.poster, m.release_date, s.date AS watched_at
FROM public.seen AS s
	INNER JOIN public.movie AS m ON m.id = s.movie_id
WHERE
	user_id = 1
	AND EXTRACT(YEAR FROM s.date) = EXTRACT(YEAR FROM CURRENT_DATE)
ORDER BY s.date DESC
`)

	if err != nil {
		panic(err)
	}

	return c.Render("index", fiber.Map{
		"Movies": movies,
	})
}

func MovieHandler(c *fiber.Ctx) error {
	var movie Movie

	err := db.Client.Get(&movie, `
SELECT
	m.*,
  ARRAY_AGG(g.name) AS genres
FROM
	public.movie AS m
	INNER JOIN public.movie_genre AS mg ON mg.movie_id = m.id
	INNER JOIN public.genre AS g ON g.id = mg.genre_id
WHERE m.id = $1
GROUP BY 1
`, c.Params("id"))

	if err != nil {
		panic(err)
	}

	return c.Render("movie", fiber.Map{
		"Movie": movie,
	})
}
