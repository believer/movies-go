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

	movieId := c.Params("id")
	userId := c.Locals("UserId")
	log.Println("id", movieId, "user", userId)
	err := db.Dot.Get(db.Client, &movie, "movie-by-id", movieId, userId)

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
	var watchlist types.Movies

	isAuth := utils.IsAuthenticated(c)
	id := c.Params("id")

	err := db.Dot.Select(db.Client, &watchedAt, "seen-by-user-id", id, c.Locals("UserId"))

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &watchlist, "is-in-watchlist", c.Locals("UserId"), id)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.Watched(components.WatchedProps{
		WatchedAt:   watchedAt,
		IsAdmin:     isAuth,
		InWatchlist: len(watchlist) > 0,
		ID:          id,
	}))
}

// Render the add movie page
func HandleGetMovieNew(c *fiber.Ctx) error {
	var watchlist types.Movies
	var movie types.Movie

	isAuth := utils.IsAuthenticated(c)
	id := c.QueryInt("id")
	imdbId := c.Query("imdbId")
	userId := c.Locals("UserId")

	if !isAuth {
		return c.Redirect("/")
	}

	if id != 0 {
		err := db.Dot.Select(db.Client, &watchlist, "is-in-watchlist", userId, id)

		if err != nil {
			return err
		}

		err = db.Client.Get(&movie, `SELECT id, title FROM movie WHERE id = $1`, id)

		if err != nil {
			return err
		}
	}

	return utils.TemplRender(c, views.NewMovie(views.NewMovieProps{
		ImdbID:      imdbId,
		InWatchlist: len(watchlist) > 0,
		Movie:       movie,
	}))
}

func tmdbFetchMovie(id string) types.MovieDetailsResponse {
	tmdbKey := os.Getenv("TMDB_API_KEY")

	resp, err := http.Get(fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?api_key=%s", id, tmdbKey))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		log.Printf("Movie information not found")
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Print(err)
	}

	var result types.MovieDetailsResponse

	err = json.Unmarshal([]byte(body), &result)

	if err != nil {
		log.Print(err)
	}

	return result
}

func tmdbFetchMovieCredits(id string) types.MovieCreditsResponse {
	tmdbKey := os.Getenv("TMDB_API_KEY")

	resp, err := http.Get(fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/credits?api_key=%s", id, tmdbKey))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		log.Fatal("Movie credits not found")
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var result types.MovieCreditsResponse

	err = json.Unmarshal([]byte(body), &result)

	if err != nil {
		log.Fatal(err)
	}

	return result
}

func tmdbSearchMovie(query string) types.SearchMovieResponse {
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

	var result types.SearchMovieResponse

	err = json.Unmarshal([]byte(body), &result)

	return result
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
		ImdbID      string `form:"imdb_id"`
		Rating      int    `form:"rating"`
		IsWatchlist bool   `form:"watchlist"`
		WatchedAt   string `form:"watched_at"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	imdbId, err := utils.ParseImdbId(data.ImdbID)

	if err != nil {
		return err
	}

	movieId := 0
	movie := tmdbFetchMovie(imdbId)
	movieCast := tmdbFetchMovieCredits(imdbId)

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
	err = tx.Get(&movieId, `INSERT INTO movie (title, runtime, release_date, imdb_id, overview, poster, tagline) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (imdb_id) DO UPDATE SET title = $1 RETURNING id`, movie.Title, movie.Runtime, movie.ReleaseDate, movie.ImdbId, movie.Overview, movie.Poster, movie.Tagline)

	if err != nil {
		return err
	}

	log.Println("Movie inserted")

	userId := c.Locals("UserId").(string)

	if data.IsWatchlist {
		// Add to watchlist
		_, err = tx.Exec(`INSERT INTO watchlist (user_id, movie_id) VALUES ($1, $2)`, userId, movieId)

		if err != nil {
			c.Set("HX-Retarget", "#error")
			return c.SendString("Movie already added to watchlist")
		}
	} else {
		// Insert a view and delete from watchlist if exists
		tx.MustExec(`INSERT INTO seen (user_id, movie_id, date) VALUES ($1, $2, $3)`, userId, movieId, watchedAt)
		tx.MustExec(`DELETE FROM watchlist WHERE user_id = $1 AND movie_id = $2`, userId, movieId)
	}

	// Insert rating
	if data.Rating != 0 {
		tx.MustExec(`INSERT INTO rating (user_id, movie_id, rating) VALUES ($1, $2, $3)`, userId, movieId, data.Rating)
	}

	type Genre struct {
		Name    string `db:"name"`
		MovieId int    `db:"movie_id"`
	}

	var genres []Genre

	// Insert genres
	for _, genre := range movie.Genres {
		genres = append(genres, Genre{
			Name:    genre.Name,
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
	for _, cast := range movieCast.Cast {
		var pfp sql.NullString
		var char sql.NullString

		if cast.ProfilePath == nil {
			pfp = sql.NullString{String: "", Valid: false}
		} else {
			pfp = sql.NullString{String: *cast.ProfilePath, Valid: true}
		}

		if cast.Character == nil {
			char = sql.NullString{String: "", Valid: false}
		} else {
			char = sql.NullString{String: *cast.Character, Valid: true}
		}

		personIndex, exists := personExists(castStructs, cast.Id, "cast")

		if exists {
			castStructs[personIndex].Name = cast.Name
			castStructs[personIndex].Popularity = cast.Popularity
			castStructs[personIndex].Character = char
			castStructs[personIndex].ProfilePicture = pfp

			continue
		}

		castStructs = append(castStructs, NewPerson{
			ID:             cast.Id,
			Name:           cast.Name,
			Popularity:     cast.Popularity,
			Character:      char,
			ProfilePicture: pfp,
			MovieId:        movieId,
		})
	}

	// Crew
	for _, crew := range movieCast.Crew {
		department := crew.Department

		if department != "Directing" && department != "Writing" && department != "Production" && department != "Sound" {
			continue
		}

		var pfp sql.NullString

		if crew.ProfilePath == nil {
			pfp = sql.NullString{String: "", Valid: false}
		} else {
			pfp = sql.NullString{String: *crew.ProfilePath, Valid: true}
		}

		if crew.Job == nil {
			continue
		}

		job := *crew.Job

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

		jobStr := sql.NullString{String: job, Valid: true}
		personIndex, exists := personExists(crewStructs, crew.Id, job)

		if exists {
			crewStructs[personIndex].Name = crew.Name
			crewStructs[personIndex].Popularity = crew.Popularity
			crewStructs[personIndex].ProfilePicture = pfp

			continue
		}

		crewStructs = append(crewStructs, NewPerson{
			ID:             crew.Id,
			Name:           crew.Name,
			Popularity:     crew.Popularity,
			Job:            jobStr,
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

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%d?back=true", movieId))

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

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%s?back=true", c.Params("id")))

	return c.SendStatus(fiber.StatusOK)
}

func HandleSearchNew(c *fiber.Ctx) error {
	query := c.Query("search")

	if query == "" {
		return utils.TemplRender(c, views.MovieSearch([]types.SearchResult{}))
	}

	movies := tmdbSearchMovie(query)

	return utils.TemplRender(c, views.MovieSearch(movies.Results))
}

func HandleGetMoviesByYear(c *fiber.Ctx) error {
	var movies []types.Movie

	year := c.Params("year")
	userId := c.Locals("UserId").(string)

	err := db.Dot.Select(db.Client, &movies, "movies-by-year", userId, year)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, views.MoviesByYear(year, movies))
}
