package routes

import (
	"believer/movies/db"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
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
	Id          int             `db:"id" json:"id"`
	ImdbId      string          `db:"imdb_id"`
	ImdbRating  sql.NullFloat64 `db:"imdb_rating"`
	Overview    string          `db:"overview"`
	Poster      string          `db:"poster"`
	ReleaseDate time.Time       `db:"release_date" json:"release_date"`
	Runtime     int             `db:"runtime"`
	Tagline     string          `db:"tagline"`
	Title       string          `db:"title" json:"title"`
	UpdatedAt   time.Time       `db:"updated_at"`
	WatchedAt   time.Time       `db:"watched_at"`
}

type Movies []Movie

func (u *Movies) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

// Format runtime in hours and minutes from minutes
func (m Movie) RuntimeFormatted() string {
	hours := m.Runtime / 60
	minutes := m.Runtime % 60

	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func FeedHandler(c *fiber.Ctx) error {
	var movies []Movie

	pageQuery := c.Query("page", "1")
	page, err := strconv.Atoi(pageQuery)

	if err != nil {
		page = 1
	}

	err = db.Client.Select(&movies, `
SELECT m.id, m.title, m.overview, m.release_date, s.date AS watched_at
FROM public.seen AS s
	INNER JOIN public.movie AS m ON m.id = s.movie_id
WHERE
	user_id = 1
ORDER BY s.date DESC
OFFSET $1
LIMIT 20
`, (page-1)*20)

	if err != nil {
		panic(err)
	}

	return c.Render("index", fiber.Map{
		"Movies":   movies,
		"NextPage": page + 1,
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

type Person struct {
	Id     int    `json:"id" db:"id"`
	Name   string `json:"name" db:"name"`
	Movies Movies `json:"movies" db:"movies"`
}

type Persons []Person

func (u *Persons) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

type Cast struct {
	Job    string  `db:"job"`
	Person Persons `db:"person"`
}

func MovieCastHandler(c *fiber.Ctx) error {

	var cast []Cast

	err := db.Client.Select(&cast, `
SELECT 
    INITCAP(mp.job::text) as job,
    JSONB_AGG(JSON_BUILD_OBJECT('name',p.name, 'id', p.id)) AS person
FROM 
    public.movie_person AS mp
    INNER JOIN public.person AS p ON p.id = mp.person_id
WHERE movie_id = $1
GROUP BY mp.job
ORDER BY
	CASE mp.job
		WHEN 'cast' THEN 1
		WHEN 'director' THEN 2
    WHEN 'composer' THEN 3
		WHEN 'writer' THEN 4
		WHEN 'producer' THEN 5
	END
`, c.Params("id"))

	if err != nil {
		panic(err)
	}

	return c.Render("partials/castList", fiber.Map{
		"Cast": cast,
	}, "")
}

func MovieSeenHandler(c *fiber.Ctx) error {
	var watchedAt []time.Time

	err := db.Client.Select(&watchedAt, `
SELECT date
FROM public.seen
WHERE movie_id = $1 AND user_id = 1
ORDER BY date DESC
`, c.Params("id"))

	if err != nil {
		panic(err)
	}

	return c.Render("partials/watched", fiber.Map{
		"WatchedAt": watchedAt,
	}, "")
}

func SearchHandler(c *fiber.Ctx) error {
	var movies []Movie

	search := c.FormValue("search")

	if search == "" {
		return c.Redirect("/")
	}

	err := db.Client.Select(&movies, `
SELECT m.id, m.title, m.overview, m.release_date, s.date AS watched_at
FROM public.seen AS s
	INNER JOIN public.movie AS m ON m.id = s.movie_id
WHERE
	user_id = 1
	AND m.title ILIKE '%' || $1 || '%'
ORDER BY s.date DESC
`, search)

	if err != nil {
		panic(err)
	}

	return c.Render("index", fiber.Map{
		"Movies": movies,
	})
}

func PersonHandler(c *fiber.Ctx) error {
	var person Person

	err := db.Client.Get(&person, `
SELECT
   p.id,
   p.name, 
   jsonb_agg(json_build_object('title', m.title, 'id', m.id, 'release_date', to_char(m.release_date, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))) as movies
FROM
  public.person as p
JOIN LATERAL (
   SELECT
   m.id,
      m.title,
      m.release_date
   FROM
      public.movie_person as mp
      INNER JOIN public.movie as m ON m.id = mp.movie_id
   WHERE
      p.id = mp.person_id
   ORDER BY m.release_date DESC
) as m ON true
WHERE p.id = $1
GROUP BY p.id, p.name;
`, c.Params("id"))

	if err != nil {
		panic(err)
	}

	return c.Render("person", fiber.Map{
		"Person": person,
	})
}
