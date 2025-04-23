package handlers

import (
	"believer/movies/components"
	"believer/movies/components/movie"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/utils/awards"
	"believer/movies/views"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

func GetMovieByID(c *fiber.Ctx) error {
	var movie types.Movie
	var review types.Review
	var isInWatchlist bool

	backParam := c.QueryBool("back", false)

	movieId := c.Params("id")
	userId := c.Locals("UserId")
	id := utils.SelfHealingUrl(movieId)
	err := db.Dot.Get(db.Client, &movie, "movie-by-id", id, userId)

	if err != nil {
		// TODO: Handle this better
		if err == sql.ErrNoRows {
			return c.Status(404).SendString("Movie not found")
		}

		return err
	}

	err = db.Dot.Get(db.Client, &review, "review-by-movie-id", id, userId)

	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	err = db.Client.Get(
		&isInWatchlist,
		`select exists (select * from watchlist where movie_id = $1 and user_id = $2);`,
		id,
		userId,
	)

	if err != nil {
		return err
	}

	if c.Get("Accept") == "application/json" {
		return c.JSON(movie)
	}

	return utils.TemplRender(c, views.Movie(
		views.MovieProps{
			IsInWatchlist: isInWatchlist,
			Movie:         movie,
			Review:        review,
			Back:          backParam,
		}))
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

func GetMovieCastByID(c *fiber.Ctx) error {
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

func GetMovieSeenByID(c *fiber.Ctx) error {
	var watchedAt []movie.WatchedAt
	var watchlist types.Movies

	isAuth := utils.IsAuthenticated(c)
	id := c.Params("id")
	imdbId := c.Query("imdbId")
	userId := c.Locals("UserId")

	err := db.Dot.Select(db.Client, &watchedAt, "seen-by-user-id", id, userId)

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &watchlist, "is-in-watchlist", userId, id)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, movie.Watched(movie.WatchedProps{
		WatchedAt:   watchedAt,
		IsAdmin:     isAuth,
		IsUnseen:    len(watchedAt) == 0,
		InWatchlist: len(watchlist) > 0,
		ImdbId:      imdbId,
		ID:          id,
	}))
}

// Render the add movie page
func GetMovieNew(c *fiber.Ctx) error {
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

func GetMovieNewSeries(c *fiber.Ctx) error {
	var options []components.DataListItem

	err := db.Client.Select(&options, `SELECT id as "value", name FROM series ORDER BY name ASC`)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.DataList(options, "series_list"))
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

	if err != nil {
		log.Fatal(err)
	}

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

func personExists(arr []NewPerson, id int, job any) (int, bool) {
	for i, person := range arr {
		if person.ID == id && person.Job.String == job {
			return i, true
		}
	}

	return 0, false
}

// Handle adding a movie
func PostMovieNew(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		HasWilhelmScream bool   `form:"wilhelm_scream"`
		ImdbID           string `form:"imdb_id"`
		IsPrivateReview  bool   `form:"review_private"`
		IsWatchlist      bool   `form:"watchlist"`
		NumberInSeries   int    `form:"number_in_series"`
		Rating           int    `form:"rating"`
		Review           string `form:"review"`
		Series           string `form:"series"`
		WatchedAt        string `form:"watched_at"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	imdbId, err := utils.ParseId(data.ImdbID)
	userId := c.Locals("UserId").(string)

	if err != nil {
		c.Set("HX-Retarget", "#error")
		return c.SendString(err.Error())
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
	err = db.Dot.Get(
		db.Client,
		&movieId,
		"insert-movie",
		movie.Title,
		movie.Runtime,
		movie.ReleaseDate,
		movie.ImdbId,
		movie.Overview,
		movie.Poster,
		movie.Tagline,
		data.HasWilhelmScream,
	)

	if err != nil {
		return err
	}

	log.Println("Movie inserted")

	// Add review if any
	if data.Review != "" {
		db.Dot.MustExec(db.Client, "insert-review", data.Review, data.IsPrivateReview, userId, movieId)
	}

	// Insert series
	if data.Series != "" && data.NumberInSeries != 0 {
		seriesId, err := strconv.Atoi(data.Series)

		// Series can't be turned into an int, so it's a new series
		if err != nil {
			err = tx.Get(&seriesId, `INSERT INTO series (name) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id`, data.Series)

			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(`INSERT INTO movie_series (movie_id, series_id, number_in_series) VALUES ($1, $2, $3)`, movieId, seriesId, data.NumberInSeries)

		if err != nil {
			c.Set("HX-Retarget", "#error")
			return c.SendString(fmt.Sprintf("Movie #%d already exists in series", data.NumberInSeries))
		}

		log.Println("Series inserted")
	}

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

	// Insert languages
	type Language struct {
		ISO639      string `db:"iso_639_1"`
		EnglishName string `db:"english_name"`
		Name        string `db:"name"`
		MovieId     int    `db:"movie_id"`
	}

	languages := make([]Language, len(movie.SpokenLanguages))

	for i, l := range movie.SpokenLanguages {
		languages[i] = Language{
			ISO639:      l.ISO639,
			Name:        l.Name,
			EnglishName: l.EnglishName,
			MovieId:     movieId,
		}
	}

	if len(languages) > 0 {
		if _, err := tx.NamedExec(
			`INSERT INTO language (name, english_name, iso_639_1) 
     VALUES (:name, :english_name, :iso_639_1) 
     ON CONFLICT DO NOTHING`, languages,
		); err != nil {
			return err
		}

		if _, err := tx.NamedExec(
			`INSERT INTO movie_language (movie_id, language_id) 
     VALUES (:movie_id, (SELECT id FROM language WHERE name = :name)) 
     ON CONFLICT DO NOTHING`, languages,
		); err != nil {
			return err
		}
	}

	// Insert genres
	type Genre struct {
		Name    string `db:"name"`
		MovieId int    `db:"movie_id"`
	}

	genres := make([]Genre, len(movie.Genres))

	for i, genre := range movie.Genres {
		genres[i] = Genre{
			Name:    genre.Name,
			MovieId: movieId,
		}
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

		personIndex, exists := personExists(castStructs, cast.ID, "cast")

		if exists {
			castStructs[personIndex].Name = cast.Name
			castStructs[personIndex].Popularity = cast.Popularity
			castStructs[personIndex].Character = char
			castStructs[personIndex].ProfilePicture = pfp

			continue
		}

		castStructs = append(castStructs, NewPerson{
			ID:             cast.ID,
			Name:           cast.Name,
			Popularity:     cast.Popularity,
			Job:            sql.NullString{String: "cast", Valid: true},
			Character:      char,
			ProfilePicture: pfp,
			MovieId:        movieId,
		})
	}

	// Crew
	for _, crew := range movieCast.Crew {
		department := crew.Department

		if department != "Directing" && department != "Writing" && department != "Production" && department != "Sound" && department != "Camera" && department != "Editing" {
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
		} else if job == "Director of Photography" {
			job = "cinematographer"
		} else if job == "Editor" {
			job = "editor"
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
			log.Println("Could not insert person")
			return err
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job, character)
	    VALUES (:movie_id, (SELECT id FROM person WHERE original_id = :id), 'cast', :character)
	ON CONFLICT (movie_id, person_id, job)
	DO UPDATE SET character = excluded.character
	`, castStructs)

		if err != nil {
			log.Println("Could not insert movie_person")
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
			log.Println("Could not insert crew")
			return err
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job)
	   VALUES (:movie_id, (SELECT id FROM person WHERE original_id = :id), :job)
	ON CONFLICT DO NOTHING
	`, crewStructs)

		if err != nil {
			log.Println("Could not insert movie_person crew")
			return err
		}
	}

	log.Println("Crew inserted")

	err = tx.Commit()

	if err != nil {
		err = tx.Rollback()

		return err
	}

	// Add awards
	awards.Add(movie.ImdbId)

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%d?back=true", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func GetByImdbId(c *fiber.Ctx) error {
	var movie types.Movie

	imdbId, err := utils.ParseId(c.Query("imdb_id"))

	if err != nil {
		return c.SendString("")
	}

	err = db.Client.Get(&movie, `SELECT id, title FROM movie WHERE imdb_id = $1`, imdbId)

	if err != nil || movie.ID == 0 {
		return c.SendString("")
	}

	return utils.TemplRender(c, components.MovieExists(movie))
}

func DeleteSeenMovie(c *fiber.Ctx) error {
	var watchedAt []movie.WatchedAt

	movieId := c.Params("id")
	seenId := c.Params("seenId")
	userId := c.Locals("UserId")
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	_, err := db.Client.Exec(`DELETE FROM seen WHERE id = $1`, seenId)

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &watchedAt, "seen-by-user-id", movieId, userId)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, movie.Watched(movie.WatchedProps{
		WatchedAt: watchedAt,
		IsAdmin:   isAuth,
		ID:        movieId,
	}))
}

func GetSeenMovie(c *fiber.Ctx) error {
	var time string

	movieId := c.Params("id")
	seenId := c.Params("seenId")

	err := db.Client.Get(&time, `SELECT TO_CHAR(date AT TIME ZONE 'UTC' AT TIME ZONE 'Europe/Stockholm', 'YYYY-MM-DD"T"HH24:MI') as date FROM seen WHERE id = $1`, seenId)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, views.UpdateWatched(views.UpdateWatchedProps{
		MovieId: movieId,
		SeenId:  seenId,
		Time:    time,
	}))
}

func UpdateSeenMovie(c *fiber.Ctx) error {
	movieId := c.Params("id")
	seenId := c.Params("seenId")
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Parse form data and watched at time
	data := new(struct {
		WatchedAt string `form:"watched_at"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
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

	_, err = db.Client.Exec(`UPDATE seen SET date = $1 WHERE id = $2`, watchedAt, seenId)

	if err != nil {
		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%s", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func CreateSeenMovie(c *fiber.Ctx) error {
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

func HandleSearch(c *fiber.Ctx) error {
	query := c.Query("search")

	if query == "" {
		return utils.TemplRender(c, views.MovieSearch([]types.SearchResult{}))
	}

	movies := tmdbSearchMovie(query)

	return utils.TemplRender(c, views.MovieSearch(movies.Results))
}

func GetMoviesByYear(c *fiber.Ctx) error {
	var movies []types.Movie

	year := c.Params("year")
	userId := c.Locals("UserId").(string)

	err := db.Dot.Select(db.Client, &movies, "movies-by-year", userId, year)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.ListView(components.ListViewProps{
		EmptyState: "No movies this year",
		Movies:     movies,
		Name:       year,
	}))
}

func DeleteRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId := c.Params("id")
	userId := c.Locals("UserId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	_, err := db.Client.Exec(`DELETE FROM rating WHERE movie_id = $1 AND user_id = $2`, movieId, userId)

	if err != nil {
		return err
	}

	return c.SendString("")
}

func GetRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")
	rating := c.QueryInt("rating")

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return utils.TemplRender(c, components.EditRating(components.EditRatingProps{
		CurrentRating: rating,
		MovieId:       movieId,
	}))
}

func GetEditRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return utils.TemplRender(c, components.AddRatingForm(components.AddRatingProps{
		MovieId: movieId,
	}))
}

func PostRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")
	userId := c.Locals("UserId")

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		Rating string `form:"rating"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	_, err = db.Client.Exec(`INSERT INTO rating (user_id, movie_id, rating) VALUES ($1, $2, $3)`, userId, movieId, data.Rating)

	if err != nil {
		return err
	}

	rating, err := strconv.ParseInt(data.Rating, 10, 64)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.Rating(components.RatingProps{
		MovieId: movieId,
		Rating:  rating,
		RatedAt: time.Now(),
	}))
}

func UpdateRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")
	userId := c.Locals("UserId")

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		Rating string `form:"rating"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	_, err = db.Client.Exec(`UPDATE rating SET rating = $1, updated_at = NOW() WHERE movie_id = $2 AND user_id = $3`, data.Rating, movieId, userId)

	if err != nil {
		return err
	}

	rating, _ := strconv.ParseInt(data.Rating, 10, 0)

	return utils.TemplRender(c, components.Rating(components.RatingProps{
		MovieId: movieId,
		Rating:  rating,
		RatedAt: time.Now(),
	}))
}

func GetMovieAwards(c *fiber.Ctx) error {
	var awards []types.Award

	imdbId := c.Params("imdbId")

	err := db.Dot.Select(db.Client, &awards, "movie-awards", imdbId)

	if err != nil {
		return err
	}

	won := 0

	for _, award := range awards {
		if award.Winner {
			won++
		}
	}

	return utils.TemplRender(c, components.MovieAwards(components.MovieAwardsProps{
		Awards: awards,
		Won:    won,
	}))
}
