package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

func HandleGetMovieByID(c *fiber.Ctx) error {
	var movie types.Movie

	err := db.Client.Get(&movie, `
SELECT
	m.*,
  ARRAY_AGG(g.name) AS genres
FROM
	movie AS m
	INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
	INNER JOIN genre AS g ON g.id = mg.genre_id
WHERE m.id = $1
GROUP BY 1
`, c.Params("id"))

	if err != nil {
		err := db.Client.Get(&movie, `
    SELECT
	m.*,
  ARRAY_AGG(g.name) AS genres
FROM
	movie AS m
	INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
	INNER JOIN genre AS g ON g.id = mg.genre_id
-- Slugify function is defined in the database
WHERE slugify(m.title) ILIKE '%' || slugify($1) || '%'
GROUP BY 1
  `, c.Params("id"))

		if err != nil {
			// TODO: Handle this better
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("Movie not found")
			}

			return err
		}
	}

	return c.Render("movie", fiber.Map{
		"Movie": movie,
	})
}

type CastDB struct {
	Job        string         `db:"job"`
	Names      pq.StringArray `db:"people_names"`
	Ids        pq.Int32Array  `db:"people_ids"`
	Characters pq.StringArray `db:"characters"`
}

type CastAndCrewDTO struct {
	Name      string
	ID        int32
	Character string
}

type CastDTO struct {
	Job    string
	People []CastAndCrewDTO
}

func ZipCast(names []string, ids []int32, characters []string) []CastAndCrewDTO {
	zipped := make([]CastAndCrewDTO, len(names))
	for i := range names {
		zipped[i] = CastAndCrewDTO{names[i], ids[i], characters[i]}
	}
	return zipped
}

func HandleGetMovieCastByID(c *fiber.Ctx) error {
	var castOrCrew []CastDB

	err := db.Dot.Select(db.Client, &castOrCrew, "cast-by-id", c.Params("id"))

	if err != nil {
		return err
	}

	updatedCastOrCrew := make([]CastDTO, len(castOrCrew))
	hasCharacters := false

	for i, cast := range castOrCrew {
		characters := cast.Characters

		if cast.Job == "Cast" {
			for _, value := range characters {
				if value != "" {
					hasCharacters = true
					break
				}
			}
		}

		if len(characters) == 0 {
			characters = make([]string, len(cast.Names))
		}

		updatedCastOrCrew[i] = CastDTO{cast.Job, ZipCast(cast.Names, cast.Ids, characters)}
	}

	return c.Render("partials/castList", fiber.Map{
		"CastOrCrew":    updatedCastOrCrew,
		"HasCharacters": hasCharacters,
	}, "")
}

func HandleGetMovieSeenByID(c *fiber.Ctx) error {
	var watchedAt []time.Time

	err := db.Client.Select(&watchedAt, `
SELECT date
FROM seen
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

// Render the add movie page
func HandleGetMovieNew(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if isAuth == false {
		return c.Redirect("/")
	}

	return c.Render("add", nil)
}

func tmdbFetchMovie(route string) map[string]interface{} {
	tmdbBaseUrl := "https://api.themoviedb.org/3/movie"
	tmdbKey := os.Getenv("TMDB_API_KEY")

	resp, err := http.Get(tmdbBaseUrl + route + "?api_key=" + tmdbKey)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(body), &result)

	return result
}

// Handle adding a movies
func HandlePostMovieNew(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if isAuth == false {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		ImdbID    string `form:"imdb_id"`
		Rating    int    `form:"rating"`
		WatchedAt string `form:"watched_at"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	imdbId, err := utils.ParseImdbId(data.ImdbID)

	if err != nil {
		return err
	}

	var (
		movieInformation = tmdbFetchMovie("/" + imdbId)
		movieCast        = tmdbFetchMovie("/" + imdbId + "/credits")
		movieId          = 0
	)

	watchedAt, err := time.Parse("2006-01-02T15:04", data.WatchedAt)

	if err != nil {
		watchedAt = time.Now()
	}

	tx := db.Client.MustBegin()

	// Insert movie information
	tx.Get(&movieId, `INSERT INTO movie (title, runtime, release_date, imdb_id, overview, poster, tagline) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (imdb_id) DO UPDATE SET title = $1 RETURNING id`, movieInformation["title"], movieInformation["runtime"], movieInformation["release_date"], imdbId, movieInformation["overview"], movieInformation["poster_path"], movieInformation["tagline"])

	// Insert a view
	tx.MustExec(`INSERT INTO seen (user_id, movie_id, date) VALUES ($1, $2, $3)`, 1, movieId, watchedAt)

	// Insert rating
	tx.MustExec(`INSERT INTO rating (user_id, movie_id, rating) VALUES ($1, $2, $3)`, 1, movieId, data.Rating)

	// Insert genres
	for _, genre := range movieInformation["genres"].([]interface{}) {
		name := genre.(map[string]interface{})["name"]

		tx.MustExec(`INSERT INTO genre (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, name)
		tx.MustExec(`INSERT INTO movie_genre (movie_id, genre_id) VALUES ($1, (SELECT id FROM genre WHERE name = $2)) ON CONFLICT DO NOTHING`, movieId, name)
	}

	// Insert cast
	for _, cast := range movieCast["cast"].([]interface{}) {
		id := cast.(map[string]interface{})["id"]
		name := cast.(map[string]interface{})["name"]
		character := cast.(map[string]interface{})["character"]
		popularity := cast.(map[string]interface{})["popularity"]
		profilePicture := cast.(map[string]interface{})["profile_path"]

		tx.MustExec(`INSERT INTO person (name, original_id, popularity, profile_picture) VALUES ($1, $2, $3, $4) ON CONFLICT (original_id) DO UPDATE SET popularity = $3, profile_picture = $4`, name, id, popularity, profilePicture)
		tx.MustExec(`INSERT INTO movie_person (movie_id, person_id, job, character) VALUES ($1, (SELECT id FROM person WHERE original_id = $4), $2, $3) ON CONFLICT (movie_id, person_id, job) DO UPDATE SET character = excluded.character`, movieId, "cast", character, id)
	}

	tx.Commit()

	c.Set("HX-Redirect", "/movies/"+fmt.Sprint(movieId))

	return c.SendStatus(fiber.StatusOK)
}

func HandleGetByImdbId(c *fiber.Ctx) error {
	var movie types.Movie

	imdbId, err := utils.ParseImdbId(c.Query("imdb_id"))

	if err != nil {
		return err
	}

	err = db.Client.Get(&movie, `
SELECT id, title FROM movie WHERE imdb_id = $1
`, imdbId)

	if err != nil || movie.ID == 0 {
		return c.SendString("")
	}

	return c.Render("partials/movieExists", fiber.Map{
		"Movie": movie,
	}, "")
}

func HandlePostMovieSeenNew(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if isAuth == false {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tx := db.Client.MustBegin()

	tx.MustExec(`INSERT INTO seen (user_id, movie_id) VALUES ($1, $2)`, 1, c.Params("id"))

	err := tx.Commit()

	if err != nil {
		tx.Rollback()
		return err
	}

	c.Set("HX-Redirect", "/movies/"+c.Params("id"))

	return c.SendStatus(fiber.StatusOK)
}
