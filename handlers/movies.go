package handlers

import (
	"believer/movies/components/list"
	"believer/movies/components/movie"
	"believer/movies/components/rating"
	"believer/movies/components/review"
	"believer/movies/components/seen"
	"believer/movies/db"
	"believer/movies/services/tmdb"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/utils/awards"
	"believer/movies/views"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

func GetMovieByID(c *fiber.Ctx) error {
	backParam := c.QueryBool("back", false)
	movieQueries, err := db.MakeMovieQueries(c)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	movie, err := movieQueries.GetByID()

	if err != nil {
		if err == sql.ErrNoRows {
			return utils.Render(c, views.NotFound())
		}

		if pgErr, ok := err.(*pq.Error); ok {
			if strings.Contains(pgErr.Message, "is out of range for type integer") {
				return utils.Render(c, views.NotFound())
			}
		}

		return err
	}

	review, err := movieQueries.ReviewByMovieID()

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	isInWatchlist, err := movieQueries.IsWatchlisted()

	if err != nil {
		return err
	}

	others, err := movieQueries.RatingsByOthers()

	if err != nil {
		return err
	}

	watchedAt, err := movieQueries.SeenByUser()

	if err != nil {
		return err
	}

	cast, hasCharacters, err := movieQueries.Cast()

	if err != nil {
		return err
	}

	if c.Get("Accept") == "application/json" {
		return c.JSON(movie)
	}

	return utils.Render(c, views.Movie(
		views.MovieProps{
			Cast:          cast,
			HasCharacters: hasCharacters,
			WatchedAt:     watchedAt,
			IsInWatchlist: isInWatchlist,
			Movie:         movie,
			Others:        others,
			Review:        review,
			Back:          backParam,
		}))
}

func GetMovieOthersSeenByID(c *fiber.Ctx) error {
	movieQueries, err := db.MakeMovieQueries(c)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	idAsInt, err := strconv.Atoi(movieQueries.Id)

	if err != nil {
		return err
	}

	others, err := movieQueries.RatingsByOthers()

	if err != nil {
		return err
	}

	return utils.Render(c, seen.MovieOthersSeen(seen.MovieOthersSeenProps{
		ID:     idAsInt,
		Others: others,
	}))
}

// Render the add movie page
func GetMovieNew(c *fiber.Ctx) error {
	var movie types.Movie

	movieQueries, err := db.MakeMovieQueries(c)
	id := c.QueryInt("id")
	imdbId := c.Query("imdbId")

	if !movieQueries.IsAuthenticated {
		return c.Redirect("/")
	}

	if err != nil {
		return utils.Render(c, views.NewMovie(views.NewMovieProps{
			ImdbID:      imdbId,
			InWatchlist: false,
			Movie:       movie,
		}))
	}

	isInWatchlist, err := movieQueries.IsWatchlisted()

	if err != nil {
		return err
	}

	err = db.Client.Get(&movie, `SELECT
    id,
    title
FROM
    movie
WHERE
    id = $1`, id)

	if err != nil {
		return err
	}

	return utils.Render(c, views.NewMovie(views.NewMovieProps{
		ImdbID:      imdbId,
		InWatchlist: isInWatchlist,
		Movie:       movie,
	}))
}

func GetMovieNewSeries(c *fiber.Ctx) error {
	var options []list.DataListItem

	err := db.Client.Select(&options, `SELECT
    id AS "value",
    name
FROM
    series
ORDER BY
    name ASC`)

	if err != nil {
		return err
	}

	return utils.Render(c, list.DataList(options, "series_list"))
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
	tmdbApi := tmdb.New(imdbId)
	movie, err := tmdbApi.Movie()

	if err != nil {
		return err
	}

	movieCast, err := tmdbApi.Credits()

	if err != nil {
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

	tx := db.Client.MustBegin()

	// Insert movie information
	err = db.Client.Get(
		&movieId,
		`
INSERT INTO movie (title, runtime, release_date, imdb_id, overview, poster, tagline, tmdb_id, wilhelm)
    VALUES ($1, $2, NULLIF ($3, '')::date, $4, $5, $6, $7, $8, $9)
ON CONFLICT (imdb_id)
    DO UPDATE SET
        title = excluded.title,
        runtime = excluded.runtime,
        release_date = excluded.release_date,
        imdb_id = excluded.imdb_id,
        overview = excluded.overview,
        poster = excluded.poster,
        tagline = excluded.tagline,
        tmdb_id = excluded.tmdb_id
    RETURNING
        id
		`,
		movie.Title,
		movie.Runtime,
		movie.ReleaseDate,
		movie.ImdbId,
		movie.Overview,
		movie.Poster,
		movie.Tagline,
		movie.TmdbId,
		data.HasWilhelmScream,
	)

	if err != nil {
		return err
	}

	log.Println("Movie inserted")

	// Add review if any
	if data.Review != "" {
		tx.MustExec(`
INSERT INTO review (content, private, user_id, movie_id)
    VALUES ($1, $2, $3, $4)
ON CONFLICT
    DO NOTHING
			`, data.Review, data.IsPrivateReview, userId, movieId)
	}

	// Insert series
	if data.Series != "" && data.NumberInSeries != 0 {
		seriesId, err := strconv.Atoi(data.Series)

		// Series can't be turned into an int, so it's a new series
		if err != nil {
			err = tx.Get(&seriesId, `INSERT INTO series (name)
    VALUES ($1)
ON CONFLICT
    DO NOTHING
RETURNING
    id`, data.Series)

			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(`INSERT INTO movie_series (movie_id, series_id, number_in_series)
    VALUES ($1, $2, $3)`, movieId, seriesId, data.NumberInSeries)

		if err != nil {
			c.Set("HX-Retarget", "#error")
			return c.SendString(fmt.Sprintf("Movie #%d already exists in series", data.NumberInSeries))
		}

		log.Println("Series inserted")
	}

	if data.IsWatchlist {
		// Add to watchlist
		_, err = tx.Exec(`INSERT INTO watchlist (user_id, movie_id)
    VALUES ($1, $2)`, userId, movieId)

		if err != nil {
			c.Set("HX-Retarget", "#error")
			return c.SendString("Movie already added to watchlist")
		}
	} else {
		// Insert a view and delete from watchlist if exists
		tx.MustExec(`INSERT INTO seen (user_id, movie_id, date)
    VALUES ($1, $2, $3)`, userId, movieId, watchedAt)
		tx.MustExec(`DELETE FROM watchlist
WHERE user_id = $1
    AND movie_id = $2`, userId, movieId)
	}

	// Insert rating
	if data.Rating != 0 {
		tx.MustExec(`INSERT INTO rating (user_id, movie_id, rating)
    VALUES ($1, $2, $3)`, userId, movieId, data.Rating)
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
			`INSERT INTO
LANGUAGE (name, english_name, iso_639_1)
    VALUES (:name, :english_name, :iso_639_1)
ON CONFLICT
    DO NOTHING`, languages,
		); err != nil {
			return err
		}

		if _, err := tx.NamedExec(
			`INSERT INTO movie_language (movie_id, language_id)
    VALUES (:movie_id, (
            SELECT
                id
            FROM
                LANGUAGE
            WHERE
                name = :name))
ON CONFLICT
    DO NOTHING`, languages,
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
		_, err = tx.NamedExec(`INSERT INTO genre (name)
    VALUES (:name)
ON CONFLICT (name)
    DO NOTHING`, genres)

		if err != nil {
			return err
		}

		_, err = tx.NamedExec(`INSERT INTO movie_genre (movie_id, genre_id)
    VALUES (:movie_id, (
            SELECT
                id
            FROM
                genre
            WHERE
                name = :name))
ON CONFLICT
    DO NOTHING`, genres)

		if err != nil {
			return err
		}
	}

	log.Println("Genres inserted")

	// Country
	for _, c := range movie.ProductionCountries {
		tx.MustExec(`
			INSERT INTO movie_country (movie_id, country_id)
			    VALUES ($1, $2)
			ON CONFLICT
			    DO NOTHING
    `, movieId, c.ID)
	}

	log.Println("Countries inserted")

	// Production companies
	for _, c := range movie.ProductionCompanies {
		tx.MustExec(`
			INSERT INTO production_company (tmdb_id, name, country)
			    VALUES ($1, $2, NULLIF ($3, ''))
			ON CONFLICT
			    DO NOTHING
		`, c.ID, c.Name, c.OriginCountry)

		tx.MustExec(`
			INSERT INTO movie_company (movie_id, company_id)
			    VALUES ($1, (
			            SELECT
			                id
			            FROM
			                production_company
			            WHERE
			                tmdb_id = $2))
			ON CONFLICT
			    DO NOTHING
		`, movieId, c.ID)
	}

	log.Println("Production companies inserted")

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

		switch job {
		case "Screenplay", "Writer", "Novel":
			job = "writer"
		case "Original Music Composer":
			job = "composer"
		case "Producer", "Associate Producer", "Executive Producer":
			job = "producer"
		case "Director":
			job = "director"
		case "Director of Photography":
			job = "cinematographer"
		case "Editor":
			job = "editor"
		default:
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
	ON CONFLICT
	    DO NOTHING
	`, castStructs)

		if err != nil {
			log.Println("Could not insert person")
			return err
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job, character)
	    VALUES (:movie_id, (
	            SELECT
	                id
	            FROM
	                person
	            WHERE
	                original_id = :id), 'cast', :character)
	ON CONFLICT (movie_id,
	    person_id,
	    job)
	    DO UPDATE SET
	        character = excluded.character
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
	ON CONFLICT
	    DO NOTHING
	`, crewStructs)

		if err != nil {
			log.Println("Could not insert crew")
			return err
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job)
	    VALUES (:movie_id, (
	            SELECT
	                id
	            FROM
	                person
	            WHERE
	                original_id = :id), :job)
	ON CONFLICT (movie_id,
	    person_id,
	    job)
	    DO UPDATE SET
	        job = excluded.job
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

	err = db.Client.Get(&movie, `SELECT
    id,
    title
FROM
    movie
WHERE
    imdb_id = $1`, imdbId)

	if err != nil || movie.ID == 0 {
		return c.SendString("")
	}

	return utils.Render(c, views.MovieExists(movie))
}

func DeleteSeenMovie(c *fiber.Ctx) error {
	q, err := db.MakeMovieQueries(c)

	if err != nil {
		return err
	}

	movieId, err := c.ParamsInt("id")
	seenId := c.Params("seenId")
	isAuth := utils.IsAuthenticated(c)

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	_, err = db.Client.Exec(`
DELETE FROM seen
WHERE id = $1
		`, seenId)

	if err != nil {
		return err
	}

	watchedAt, err := q.SeenByUser()

	if err != nil {
		return err
	}

	return utils.Render(c, movie.Watched(movie.WatchedProps{
		WatchedAt: watchedAt,
		ID:        movieId,
	}))
}

func GetSeenMovie(c *fiber.Ctx) error {
	var time string

	movieId, err := c.ParamsInt("id")
	seenId := c.Params("seenId")

	if err != nil {
		return err
	}

	err = db.Client.Get(&time, `SELECT
    TO_CHAR(date AT TIME ZONE 'UTC' AT TIME ZONE 'Europe/Stockholm', 'YYYY-MM-DD"T"HH24:MI') AS date
FROM
    seen
WHERE
    id = $1`, seenId)

	if err != nil {
		return err
	}

	return utils.Render(c, views.UpdateWatched(views.UpdateWatchedProps{
		MovieId: movieId,
		SeenId:  seenId,
		Time:    time,
	}))
}

func UpdateSeenMovie(c *fiber.Ctx) error {
	movieId, err := c.ParamsInt("id")
	seenId := c.Params("seenId")
	isAuth := utils.IsAuthenticated(c)

	if err != nil {
		return err
	}

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

	_, err = db.Client.Exec(`UPDATE
    seen
SET
    date = $1
WHERE
    id = $2`, watchedAt, seenId)

	if err != nil {
		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%d", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func CreateSeenMovie(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tx := db.Client.MustBegin()

	tx.MustExec(`INSERT INTO seen (user_id, movie_id)
    VALUES ($1, $2)`, c.Locals("UserId"), c.Params("id"))

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
		return utils.Render(c, views.MovieSearch([]types.SearchResult{}))
	}

	tmdbApi := tmdb.New("")
	movies, err := tmdbApi.Search(query)

	if err != nil {
		return err
	}

	if len(movies.Results) == 0 {
		return utils.Render(c, views.MovieSearchEmpty())
	}

	maxResults := clamp(len(movies.Results), 1, 5)

	return utils.Render(c, views.MovieSearch(movies.Results[:maxResults]))
}

func GetMoviesByYear(c *fiber.Ctx) error {
	var movies []types.Movie

	year := c.Params("year")
	userId := c.Locals("UserId").(string)

	err := db.Client.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.release_date,
    m.imdb_id,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie AS m
    LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            public.seen
        WHERE
            user_id = $1
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id
WHERE
    date_part('year', release_date) = $2
ORDER BY
    release_date
		`, userId, year)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies this year",
		Movies:     movies,
		Name:       year,
	}))
}

func DeleteRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")
	userId := c.Locals("UserId")

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	_, err = db.Client.Exec(`DELETE FROM rating
WHERE movie_id = $1
    AND user_id = $2`, movieId, userId)

	if err != nil {
		return err
	}

	return utils.Render(c, rating.AddRating(rating.AddRatingProps{
		MovieId: movieId,
	}))
}

func GetRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")
	currentRating := c.QueryInt("rating")

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return utils.Render(c, rating.EditRating(rating.EditRatingProps{
		CurrentRating: currentRating,
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

	return utils.Render(c, rating.AddRatingForm(rating.AddRatingProps{
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

	_, err = db.Client.Exec(`INSERT INTO rating (user_id, movie_id, rating)
    VALUES ($1, $2, $3)`, userId, movieId, data.Rating)

	if err != nil {
		return err
	}

	movieRating, err := strconv.ParseInt(data.Rating, 10, 64)

	if err != nil {
		return err
	}

	return utils.Render(c, rating.GetRating(rating.Props{
		MovieId: movieId,
		Rating:  movieRating,
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

	_, err = db.Client.Exec(`UPDATE
    rating
SET
    rating = $1,
    updated_at = NOW()
WHERE
    movie_id = $2
    AND user_id = $3`, data.Rating, movieId, userId)

	if err != nil {
		return err
	}

	movieRating, _ := strconv.ParseInt(data.Rating, 10, 0)

	return utils.Render(c, rating.GetRating(rating.Props{
		MovieId: movieId,
		Rating:  movieRating,
		RatedAt: time.Now(),
	}))
}

func GetMovieAwards(c *fiber.Ctx) error {
	var awards []types.Award
	var year string

	imdbId := c.Params("imdbId")

	err := db.Client.Select(&awards, `
SELECT
    name AS category,
    year,
    COALESCE(JSONB_AGG(
            CASE WHEN person IS NOT NULL
                AND person_id IS NOT NULL THEN
                JSONB_BUILD_OBJECT('name', person, 'id', person_id)
            WHEN person IS NOT NULL THEN
                JSONB_BUILD_OBJECT('name', person)
            ELSE
                JSONB_BUILD_OBJECT('name', 'N/A')
            END) FILTER (WHERE person IS NOT NULL
            OR person_id IS NOT NULL), '[]'::jsonb) AS nominees,
    winner,
    detail
FROM
    award
WHERE
    imdb_id = $1
GROUP BY
    name,
    year,
    winner,
    detail
ORDER BY
    winner DESC,
    category ASC
		`, imdbId)

	if err != nil {
		return err
	}

	won := 0

	for _, award := range awards {
		year = award.Year

		if award.Winner {
			won++
		}
	}

	return utils.Render(c, views.MovieAwards(views.MovieAwardsProps{
		Awards: awards,
		Year:   year,
		Won:    won,
	}))
}

func EditMovieReview(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	q, err := db.MakeMovieQueries(c)

	if err != nil {
		return err
	}

	reviewData, err := q.ReviewByID()

	if err != nil {
		return err
	}

	return utils.Render(c, review.EditReview(reviewData))
}

func UpdateMovieReview(c *fiber.Ctx) error {
	q, err := db.MakeMovieQueries(c)

	if err != nil {
		return err
	}

	id := c.Params("id")
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		Review          string `form:"review"`
		IsPrivateReview bool   `form:"review_private"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	_, err = db.Client.Exec(`UPDATE
    review
SET
    content = $1,
    private = $2
WHERE
    id = $3`, data.Review, data.IsPrivateReview, id)

	if err != nil {
		return err
	}

	reviewData, err := q.ReviewByID()

	if err != nil {
		return err
	}

	return utils.Render(c, review.ReviewContent(reviewData))
}

func UpdateMovieByID(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userId := c.Locals("UserId").(string)

	if userId != "1" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	id, err := c.ParamsInt("id")

	if err != nil {
		return err
	}

	imdbId := ""

	err = db.Client.Get(&imdbId, "SELECT imdb_id FROM movie WHERE id = $1", id)

	if err != nil {
		return err
	}

	tmdbApi := tmdb.New(imdbId)
	movie, err := tmdbApi.Movie()

	if err != nil {
		return err
	}

	movieCast, err := tmdbApi.Credits()

	if err != nil {
		return err
	}

	tx := db.Client.MustBegin()

	fmt.Println(
		id,
		movie.Title,
		movie.Runtime,
		movie.ReleaseDate,
		movie.ImdbId,
		movie.Overview,
		movie.Poster,
		movie.Tagline,
	)

	// Update movie information
	_, err = tx.Exec(`
	UPDATE
	    movie
	SET
	    title = $2,
	    runtime = $3,
	    release_date = NULLIF ($4, '')::date,
	    imdb_id = $5,
	    overview = $6,
	    poster = $7,
	    tagline = $8,
	    updated_at = NOW(),
	    tmdb_id = $9
	WHERE
	    id = $1;

`,
		id,
		movie.Title,
		movie.Runtime,
		movie.ReleaseDate,
		movie.ImdbId,
		movie.Overview,
		movie.Poster,
		movie.Tagline,
		movie.TmdbId,
	)

	if err != nil {
		return err
	}

	log.Println("Movie updated")

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
			MovieId:     id,
		}
	}

	if len(languages) > 0 {
		if _, err := tx.NamedExec(
			`INSERT INTO
LANGUAGE (name, english_name, iso_639_1)
    VALUES (:name, :english_name, :iso_639_1)
ON CONFLICT
    DO NOTHING`, languages,
		); err != nil {
			return err
		}

		if _, err := tx.NamedExec(
			`INSERT INTO movie_language (movie_id, language_id)
    VALUES (:movie_id, (
            SELECT
                id
            FROM
                LANGUAGE
            WHERE
                name = :name))
ON CONFLICT
    DO NOTHING`, languages,
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
			MovieId: id,
		}
	}

	if len(genres) > 0 {
		_, err = tx.NamedExec(`INSERT INTO genre (name)
    VALUES (:name)
ON CONFLICT (name)
    DO NOTHING`, genres)

		if err != nil {
			return err
		}

		_, err = tx.NamedExec(`INSERT INTO movie_genre (movie_id, genre_id)
    VALUES (:movie_id, (
            SELECT
                id
            FROM
                genre
            WHERE
                name = :name))
ON CONFLICT
    DO NOTHING`, genres)

		if err != nil {
			return err
		}
	}

	log.Println("Genres updated")

	for _, c := range movie.ProductionCountries {
		tx.MustExec(`
			INSERT INTO movie_country (movie_id, country_id)
			    VALUES ($1, $2)
			ON CONFLICT
			    DO NOTHING
    `, id, c.ID)
	}

	log.Println("Countries inserted")

	// Production companies
	for _, c := range movie.ProductionCompanies {
		tx.MustExec(`
			INSERT INTO production_company (tmdb_id, name, country)
			    VALUES ($1, $2, $3)
			ON CONFLICT
			    DO NOTHING
		`, c.ID, c.Name, c.OriginCountry)

		tx.MustExec(`
			INSERT INTO movie_company (movie_id, company_id)
			    VALUES ($1, (
			            SELECT
			                id
			            FROM
			                production_company
			            WHERE
			                tmdb_id = $2))
			ON CONFLICT
			    DO NOTHING
		`, id, c.ID)
	}

	log.Println("Production companies inserted")

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
			MovieId:        id,
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

		switch job {
		case "Screenplay", "Writer", "Novel":
			job = "writer"
		case "Original Music Composer":
			job = "composer"
		case "Producer", "Associate Producer", "Executive Producer":
			job = "producer"
		case "Director":
			job = "director"
		case "Director of Photography":
			job = "cinematographer"
		case "Editor":
			job = "editor"
		default:
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
			MovieId:        id,
		})
	}

	if len(castStructs) > 0 {
		_, err = tx.NamedExec(`
	INSERT INTO person (name, original_id, popularity, profile_picture)
	    VALUES (:name, :id, :popularity, :profile_picture)
	ON CONFLICT
	    DO NOTHING
	`, castStructs)

		if err != nil {
			log.Println("Could not insert person")
			return err
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job, character)
	    VALUES (:movie_id, (
	            SELECT
	                id
	            FROM
	                person
	            WHERE
	                original_id = :id), 'cast', :character)
	ON CONFLICT (movie_id,
	    person_id,
	    job)
	    DO UPDATE SET
	        character = excluded.character
	`, castStructs)

		if err != nil {
			log.Println("Could not insert movie_person")
			return err
		}
	}

	log.Println("Cast updated")

	if len(crewStructs) > 0 {
		_, err = tx.NamedExec(`
	INSERT INTO person (name, original_id, popularity, profile_picture)
	    VALUES (:name, :id, :popularity, :profile_picture)
	ON CONFLICT
	    DO NOTHING
	`, crewStructs)

		if err != nil {
			log.Println("Could not insert crew")
			return err
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job)
	    VALUES (:movie_id, (
	            SELECT
	                id
	            FROM
	                person
	            WHERE
	                original_id = :id), :job)
	ON CONFLICT (movie_id,
	    person_id,
	    job)
	    DO UPDATE SET
	        job = excluded.job
	`, crewStructs)

		if err != nil {
			log.Println("Could not insert movie_person crew")
			return err
		}
	}

	log.Println("Crew updated")

	err = tx.Commit()

	if err != nil {
		err = tx.Rollback()

		return err
	}

	// Add awards
	awards.Add(movie.ImdbId)

	return c.SendStatus(fiber.StatusOK)

}

func WatchProviders(c *fiber.Ctx) error {
	movieQueries, err := db.MakeMovieQueries(c)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var storedProviders string
	userId := c.Locals("UserId")

	err = db.Client.Get(&storedProviders, `SELECT
    watch_providers
FROM
    "user"
WHERE
    id = $1`, userId)

	var movie types.Movie

	err = db.Client.Get(&movie, `SELECT
    imdb_id,
    title
FROM
    movie
WHERE
    id = $1`, movieQueries.Id)

	if err != nil {
		return err
	}

	t := tmdb.New(movie.ImdbId)
	watchProviders, err := t.WatchProviders()

	if err != nil {
		return err
	}

	hasProviders := len(watchProviders.Results.SE.Buy) > 0 || len(watchProviders.Results.SE.Rent) > 0 || len(watchProviders.Results.SE.Subscription) > 0

	if !hasProviders {
		return c.SendStatus(fiber.StatusNotFound)
	}

	// Only Swedish providers supported
	justWatchUrl := fmt.Sprintf("https://www.justwatch.com/se/film/%s", utils.Slugify(movie.Title))

	var myProviders views.WatchProviderCategories
	var otherProviders views.WatchProviderCategories
	hasOtherProviders := false
	hasMyProviders := false

	for _, w := range watchProviders.Results.SE.Ads {
		if strings.Contains(storedProviders, w.Name) {
			myProviders.Ads = append(myProviders.Ads, w)
			hasMyProviders = true
		} else {
			otherProviders.Ads = append(otherProviders.Ads, w)
			hasOtherProviders = true
		}
	}

	for _, w := range watchProviders.Results.SE.Buy {
		if strings.Contains(storedProviders, w.Name) {
			myProviders.Buy = append(myProviders.Buy, w)
			hasMyProviders = true
		} else {
			otherProviders.Buy = append(otherProviders.Buy, w)
			hasOtherProviders = true
		}
	}

	for _, w := range watchProviders.Results.SE.Free {
		if strings.Contains(storedProviders, w.Name) {
			myProviders.Free = append(myProviders.Free, w)
			hasMyProviders = true
		} else {
			otherProviders.Free = append(otherProviders.Free, w)
			hasOtherProviders = true
		}
	}

	for _, w := range watchProviders.Results.SE.Rent {
		if strings.Contains(storedProviders, w.Name) {
			myProviders.Rent = append(myProviders.Rent, w)
			hasMyProviders = true
		} else {
			otherProviders.Rent = append(otherProviders.Rent, w)
			hasOtherProviders = true
		}
	}

	for _, w := range watchProviders.Results.SE.Subscription {
		if strings.Contains(storedProviders, w.Name) {
			myProviders.Subscription = append(myProviders.Subscription, w)
			hasMyProviders = true
		} else {
			otherProviders.Subscription = append(otherProviders.Subscription, w)
			hasOtherProviders = true
		}
	}

	return utils.Render(c, views.WatchProviders(views.WatchProvidersProps{
		HasMyProviders:    hasMyProviders,
		HasOtherProviders: hasOtherProviders,
		OtherProviders:    otherProviders,
		MyProviders:       myProviders,
		JustWatchLink:     justWatchUrl,
	}))
}
