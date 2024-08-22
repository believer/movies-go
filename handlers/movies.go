package handlers

import (
	"believer/movies/components"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

func HandleGetMovieByID(c *fiber.Ctx) error {
	var movie types.Movie

	backParam := c.QueryBool("back", false)

	err := db.Dot.Get(db.Client, &movie, "movie-by-id", c.Params("id"), c.Locals("UserId"))

	if err != nil {
		err := db.Dot.Get(db.Client, &movie, "movie-by-name", c.Params("id"))

		if err != nil {
			// TODO: Handle this better
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("Movie not found")
			}

			return err
		}
	}

	return utils.TemplRender(c, views.Movie(movie, backParam))
}

type CastDB struct {
	Job        string         `db:"job"`
	Names      pq.StringArray `db:"people_names"`
	Ids        pq.Int32Array  `db:"people_ids"`
	Characters pq.StringArray `db:"characters"`
}

func ZipCast(names []string, ids []int32, characters []string) []components.CastAndCrewDTO {
	zipped := make([]components.CastAndCrewDTO, len(names))
	for i := range names {
		zipped[i] = components.CastAndCrewDTO{
			Name:      names[i],
			ID:        ids[i],
			Character: characters[i],
		}
	}
	return zipped
}

func HandleGetMovieCastByID(c *fiber.Ctx) error {
	var castOrCrew []CastDB

	err := db.Dot.Select(db.Client, &castOrCrew, "cast-by-id", c.Params("id"))

	if err != nil {
		return err
	}

	updatedCastOrCrew := make([]components.CastDTO, len(castOrCrew))
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

		updatedCastOrCrew[i] = components.CastDTO{
			Job:    cast.Job,
			People: ZipCast(cast.Names, cast.Ids, characters),
		}
	}

	return utils.TemplRender(c, components.CastList(updatedCastOrCrew, hasCharacters))
}

func HandleGetMovieSeenByID(c *fiber.Ctx) error {
	var watchedAt []time.Time

	isAuth := utils.IsAuthenticated(c)
	id := c.Params("id")

	err := db.Dot.Select(db.Client, &watchedAt, "seen-by-user-id", id, c.Locals("UserId"))

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.Watched(watchedAt, isAuth, id))
}

// Render the add movie page
func HandleGetMovieNew(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.Redirect("/")
	}

	return utils.TemplRender(c, views.NewMovie())
}

func tmdbFetchMovie(route string) (map[string]interface{}, error) {
	tmdbBaseUrl := "https://api.themoviedb.org/3/movie"
	tmdbKey := os.Getenv("TMDB_API_KEY")

	resp, err := http.Get(tmdbBaseUrl + route + "?api_key=" + tmdbKey)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("movie not found")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result map[string]interface{}

	err = json.Unmarshal([]byte(body), &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func tmdbSearchMovie(query string) (types.Response, error) {
	tmdbBaseUrl := "https://api.themoviedb.org/3/search/movie"
	tmdbKey := os.Getenv("TMDB_API_KEY")

	resp, err := http.Get(tmdbBaseUrl + "?query=" + url.QueryEscape(query) + "&api_key=" + tmdbKey)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result types.Response

	err = json.Unmarshal([]byte(body), &result)

	return result, nil
}

type NewPerson struct {
	ID             int            `db:"id"`
	Name           string         `db:"name"`
	Job            sql.NullString `db:"job"`
	Character      sql.NullString `db:"character"`
	Popularity     float64        `db:"popularity"`
	ProfilePicture sql.NullString `db:"profile_picture"`
	MovieId        int            `db:"movie_id"`
}

func personExists(arr []NewPerson, id int, job interface{}) (int, bool) {
	for i, person := range arr {
		if person.ID == id && person.Job == job {
			return i, true
		}
	}

	return 0, false
}

// Handle adding a movie
func HandlePostMovieNew(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
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

	movieId := 0
	movieInformation, err := tmdbFetchMovie("/" + imdbId)

	if err != nil {
		log.Println("Error fetching movie information")
	}

	movieCast, err := tmdbFetchMovie("/" + imdbId + "/credits")

	if err != nil {
		log.Println("Error fetching movie cast")
	}

	watchedAt, err := time.Parse("2006-01-02T15:04", data.WatchedAt)

	if err != nil {
		now := time.Now()
		watchedAt, err = time.Parse("2006-01-02", data.WatchedAt)

		if err != nil {
			watchedAt = now
		}

		// Set the current time
		watchedAt = watchedAt.Add(time.Duration(now.Hour()))
	}

	tx := db.Client.MustBegin()

	// Insert movie information
	err = tx.Get(&movieId, `INSERT INTO movie (title, runtime, release_date, imdb_id, overview, poster, tagline) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (imdb_id) DO UPDATE SET title = $1 RETURNING id`, movieInformation["title"], movieInformation["runtime"], movieInformation["release_date"], imdbId, movieInformation["overview"], movieInformation["poster_path"], movieInformation["tagline"])

	if err != nil {
		return err
	}

	log.Println("Movie inserted")

	userId := c.Locals("UserId").(string)

	// Insert a view
	tx.MustExec(`INSERT INTO seen (user_id, movie_id, date) VALUES ($1, $2, $3)`, userId, movieId, watchedAt)

	// Insert rating
	tx.MustExec(`INSERT INTO rating (user_id, movie_id, rating) VALUES ($1, $2, $3)`, userId, movieId, data.Rating)

	type Genre struct {
		Name    string `db:"name"`
		MovieId int    `db:"movie_id"`
	}

	var genres []Genre

	// Insert genres
	for _, genre := range movieInformation["genres"].([]interface{}) {
		name := genre.(map[string]interface{})["name"]

		genres = append(genres, Genre{
			Name:    name.(string),
			MovieId: movieId,
		})
	}

	if len(genres) > 0 {
		_, err = tx.NamedExec(`INSERT INTO genre (name) VALUES (:name) ON CONFLICT (name) DO NOTHING`, genres)

		if err != nil {
			return err
		}

		_, err = tx.NamedExec(`INSERT INTO movie_genre (movie_id, genre_id) VALUES (:movie_id, (SELECT id FROM genre WHERE name = :name)) ON CONFLICT DO NOTHING`, genres)

		if err != nil {
			return err
		}
	}

	log.Println("Genres inserted")

	var castStructs []NewPerson
	var crewStructs []NewPerson

	// Insert cast
	for _, cast := range movieCast["cast"].([]interface{}) {
		id := cast.(map[string]interface{})["id"]
		name := cast.(map[string]interface{})["name"]
		character := cast.(map[string]interface{})["character"]
		popularity := cast.(map[string]interface{})["popularity"]
		profilePicture := cast.(map[string]interface{})["profile_path"]

		var pfp sql.NullString
		var char sql.NullString

		if profilePicture == nil {
			pfp = sql.NullString{String: "", Valid: false}
		} else {
			pfp = sql.NullString{String: profilePicture.(string), Valid: true}
		}

		if character == nil {
			char = sql.NullString{String: "", Valid: false}
		} else {
			char = sql.NullString{String: character.(string), Valid: true}
		}

		idInt := int(id.(float64))

		personIndex, exists := personExists(castStructs, idInt, "cast")

		if exists {
			castStructs[personIndex].Name = name.(string)
			castStructs[personIndex].Character = char
			castStructs[personIndex].Popularity = popularity.(float64)
			castStructs[personIndex].ProfilePicture = pfp

			continue
		}

		castStructs = append(castStructs, NewPerson{
			ID:             idInt,
			Name:           name.(string),
			Character:      char,
			Popularity:     popularity.(float64),
			ProfilePicture: pfp,
			MovieId:        movieId,
		})
	}

	// Crew
	for _, crew := range movieCast["crew"].([]interface{}) {
		department := crew.(map[string]interface{})["department"]

		if department != "Directing" && department != "Writing" && department != "Production" && department != "Sound" {
			continue
		}

		job := crew.(map[string]interface{})["job"]

		if job == "Screenplay" || job == "Writer" || job == "Novel" {
			job = "writer"
		} else if job == "Original Music Composer" {
			job = "composer"
		} else if job == "Producer" || job == "Associate Producer" || job == "Executive Producer" {
			job = "producer"
		} else if job == "Director" {
			job = "director"
		} else {
			continue
		}

		id := crew.(map[string]interface{})["id"]
		name := crew.(map[string]interface{})["name"]
		popularity := crew.(map[string]interface{})["popularity"]
		profilePicture := crew.(map[string]interface{})["profile_path"]

		var pfp sql.NullString
		var jobStr sql.NullString

		if profilePicture == nil {
			pfp = sql.NullString{String: "", Valid: false}
		} else {
			pfp = sql.NullString{String: profilePicture.(string), Valid: true}
		}

		if job == nil {
			jobStr = sql.NullString{String: "", Valid: false}
		} else {
			jobStr = sql.NullString{String: job.(string), Valid: true}
		}

		idInt := int(id.(float64))

		personIndex, exists := personExists(crewStructs, idInt, job)

		if exists {
			crewStructs[personIndex].Name = name.(string)
			crewStructs[personIndex].Popularity = popularity.(float64)
			crewStructs[personIndex].ProfilePicture = pfp

			continue
		}

		crewStructs = append(crewStructs, NewPerson{
			ID:             idInt,
			Name:           name.(string),
			Job:            jobStr,
			Popularity:     popularity.(float64),
			ProfilePicture: pfp,
			MovieId:        movieId,
		})
	}

	if len(castStructs) > 0 {
		_, err = tx.NamedExec(`
	INSERT INTO person (name, original_id, popularity, profile_picture)
	VALUES (:name, :id, :popularity, :profile_picture)
	ON CONFLICT (original_id)
	DO UPDATE SET
	  popularity = excluded.popularity,
	  profile_picture = excluded.profile_picture
	`, castStructs)

		if err != nil {
			return err
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job, character)
	    VALUES (:movie_id, (SELECT id FROM person WHERE original_id = :id), 'cast', :character)
	ON CONFLICT (movie_id, person_id, job)
	DO UPDATE SET character = excluded.character
	`, castStructs)

		if err != nil {
			return err
		}
	}

	log.Println("Cast inserted")

	if len(crewStructs) > 0 {
		_, err = tx.NamedExec(`
	INSERT INTO person (name, original_id, popularity, profile_picture)
	VALUES (:name, :id, :popularity, :profile_picture)
	ON CONFLICT DO NOTHING
	`, crewStructs)

		if err != nil {
			return err
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job)
    VALUES (:movie_id, (SELECT id FROM person WHERE original_id = :id), :job)
	ON CONFLICT DO NOTHING
	`, crewStructs)

		if err != nil {
			return err
		}
	}

	log.Println("Crew inserted")

	err = tx.Commit()

	if err != nil {
		err = tx.Rollback()

		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movies/%d", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func HandleGetByImdbId(c *fiber.Ctx) error {
	var movie types.Movie

	imdbId, err := utils.ParseImdbId(c.Query("imdb_id"))

	if err != nil {
		return c.SendString("")
	}

	err = db.Client.Get(&movie, `SELECT id, title FROM movie WHERE imdb_id = $1`, imdbId)

	if err != nil || movie.ID == 0 {
		return c.SendString("")
	}

	return utils.TemplRender(c, components.MovieExists(movie))
}

func HandlePostMovieSeenNew(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tx := db.Client.MustBegin()

	tx.MustExec(`INSERT INTO seen (user_id, movie_id) VALUES ($1, $2)`, c.Locals("UserId"), c.Params("id"))

	err := tx.Commit()

	if err != nil {
		err = tx.Rollback()

		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movies/%s?back=true", c.Params("id")))

	return c.SendStatus(fiber.StatusOK)
}

func HandleSearchNew(c *fiber.Ctx) error {
	query := c.Query("search")

	if query == "" {
		return utils.TemplRender(c, views.MovieSearch([]types.SearchResult{}))
	}

	movies, err := tmdbSearchMovie(query)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, views.MovieSearch(movies.Results))
}
