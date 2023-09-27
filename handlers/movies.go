package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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

func HandleGetMovieCastByID(c *fiber.Ctx) error {
	var cast []types.Cast

	err := db.Client.Select(&cast, `
SELECT 
    INITCAP(mp.job::text) as job,
    JSONB_AGG(JSON_BUILD_OBJECT('name',p.name, 'id', p.id)) AS person
FROM 
    movie_person AS mp
    INNER JOIN person AS p ON p.id = mp.person_id
WHERE movie_id = $1
GROUP BY mp.job
ORDER BY
	CASE mp.job
		WHEN 'director' THEN 1
		WHEN 'writer' THEN 2
		WHEN 'cast' THEN 3
    WHEN 'composer' THEN 4
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
	return c.Render("add", nil)
}

func parseImdbId(s string) string {
	parsedUrl, err := url.Parse(s)

	if err != nil {
		return ""
	}

	imdbId := path.Base(parsedUrl.Path)
	imdbId = strings.TrimRight(imdbId, "/")

	return imdbId
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
	data := new(struct {
		ImdbID    string `form:"imdb_id"`
		Rating    int    `form:"rating"`
		WatchedAt string `form:"watched_at"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	imdbId := parseImdbId(data.ImdbID)
	// The returned data looks like
	// map[
	// adult:false
	// backdrop_path:/kXfqcdQKsToO0OUXHcrrNCHDBzO.jpg
	// belongs_to_collection:<nil>
	// budget:2.5e+07
	// genres:[map[id:18 name:Drama]
	// map[id:80 name:Crime]]
	// homepage:
	// id:278
	// imdb_id:tt0111161
	// original_language:en
	// original_title:The Shawshank Redemption
	// overview:Framed in the 1940s for the double murder of his wife and her lover, upstanding banker Andy Dufresne begins a new life at the Shawshank prison, where he puts his accounting skills to work for an amoral warden. During his long stretch in prison, Dufresne comes to be admired by the other inmates -- including an older prisoner named Red -- for his integrity and unquenchable sense of hope.
	// popularity:104.343
	// poster_path:/lyQBXzOQSuE59IsHyhrp0qIiPAz.jpg
	// production_companies:[map[id:97 logo_path:/qv3ih9pR9w2XNKZDsqDqAGuZjqc.pngname:Castle Rock Entertainment origin_country:US]]
	// production_countries:[map[iso_3166_1:US name:United States of America]]
	// release_date:1994-09-23
	// revenue:2.8341469e+07
	// runtime:142
	// spoken_languages:[map[english_name:English iso_639_1:en name:English]]
	// status:Released
	// tagline:Fear can hold you prisoner. Hope can set you free.
	// title:The Shawshank Redemption
	// video:false
	// vote_average:8.702
	// vote_count:24642
	// ]
	movieInformation := tmdbFetchMovie("/" + imdbId)
	// [{
	// "cast": [
	// {
	//   "adult": false,
	//   "gender": 2,
	//   "id": 819,
	//   "known_for_department": "Acting",
	//   "name": "Edward Norton",
	//   "original_name": "Edward Norton",
	//   "popularity": 26.99,
	//   "profile_path": "/8nytsqL59SFJTVYVrN72k6qkGgJ.jpg",
	//   "cast_id": 4,
	//   "character": "The Narrator",
	//   "credit_id": "52fe4250c3a36847f80149f3",
	//   "order": 0
	// },
	// ],
	// "crew": [
	// {
	//   "adult": false,
	//   "gender": 2,
	//   "id": 376,
	//   "known_for_department": "Production",
	//   "name": "Arnon Milchan",
	//   "original_name": "Arnon Milchan",
	//   "popularity": 2.931,
	//   "profile_path": "/b2hBExX4NnczNAnLuTBF4kmNhZm.jpg",
	//   "credit_id": "55731b8192514111610027d7",
	//   "department": "Production",
	//   "job": "Executive Producer"
	// },
	// ]
	// }]
	movieCast := tmdbFetchMovie("/" + imdbId + "/credits")

	movieId := 0

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

	imdbId := parseImdbId(c.Query("imdb_id"))

	err := db.Client.Get(&movie, `
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
