package handlers

import (
	"believer/movies/components/list"
	"believer/movies/components/movie"
	"believer/movies/components/rating"
	"believer/movies/components/review"
	"believer/movies/components/seen"
	"believer/movies/db"
	"believer/movies/services/api"
	"believer/movies/services/tmdb"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/utils/awards"
	"believer/movies/views"
	"database/sql"
	"fmt"
	"log/slog"
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

// Handle adding a movie
func PostMovieNew(c *fiber.Ctx) error {
	if c.Locals("IsAuthenticated") == false {
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

	api := api.New(c)

	movie, movieId, err := api.AddMovie(imdbId, data.HasWilhelmScream)
	if err != nil {
		return err
	}

	slog.Debug("Movie inserted")

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

		slog.Debug("Series inserted")
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

	// Remove from now playing if exists
	tx.MustExec(`DELETE FROM now_playing
WHERE user_id = $1
    AND imdb_id = $2`, userId, movie.ImdbId)

	// Insert rating
	if data.Rating != 0 {
		tx.MustExec(`INSERT INTO rating (user_id, movie_id, rating)
    VALUES ($1, $2, $3)`, userId, movieId, data.Rating)
	}

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
DELETE FROM seen_with
WHERE seen_id = $1
		`, seenId)

	if err != nil {
		return err
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
	var watch views.WatchData

	movieId, err := c.ParamsInt("id")
	seenId := c.Params("seenId")
	userId := c.Locals("UserId")

	if err != nil {
		return err
	}

	err = db.Client.Get(&watch, `
SELECT
    TO_CHAR(date AT TIME ZONE 'UTC' AT TIME ZONE 'Europe/Stockholm', 'YYYY-MM-DD"T"HH24:MI') AS date,
    COALESCE(ARRAY_AGG(sw.other_user_id) FILTER (WHERE sw.other_user_id IS NOT NULL), '{}') AS seen_with
FROM
    seen s
    LEFT JOIN seen_with sw ON sw.seen_id = s.id
WHERE
    id = $1
GROUP BY
    id
`, seenId)

	if err != nil {
		return err
	}

	var friends []list.DataListItem
	err = db.Client.Select(&friends, `
		SELECT
		    id AS "value",
		    name
		FROM
		    "user"
		WHERE
		    id != $1
		`, userId)

	if err != nil {
		return err
	}

	return utils.Render(c, views.UpdateWatched(views.UpdateWatchedProps{
		Friends: friends,
		MovieId: movieId,
		SeenId:  seenId,
		Watch:   watch,
	}))
}

func UpdateSeenMovie(c *fiber.Ctx) error {
	movieId, err := c.ParamsInt("id")
	seenID := c.Params("seenId")
	isAuth := utils.IsAuthenticated(c)

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Parse form data and watched at time
	data := new(struct {
		Friend    []string `form:"friend"`
		WatchedAt string   `form:"watched_at"`
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
    id = $2`, watchedAt, seenID)

	if err != nil {
		return err
	}

	_, err = db.Client.Exec(`DELETE FROM seen_with
WHERE seen_id = $1`, seenID)

	if err != nil {
		return err
	}

	if len(data.Friend) > 0 {
		_, err = db.Client.Exec(`
			INSERT INTO seen_with (seen_id, other_user_id)
			SELECT
			    $1,
			    UNNEST($2::text[])::int
			ON CONFLICT
			    DO NOTHING
			`, seenID, pq.Array(data.Friend))

		if err != nil {
			return err
		}
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
	page := c.QueryInt("page", 1)

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
    release_date ASC OFFSET $3
LIMIT 50
		`, userId, year, (page-1)*50)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies this year",
		NextPage:   fmt.Sprintf("/year/%s?page=%d", year, page+1),
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
	movieId := c.QueryInt("movieId")

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

	return utils.Render(c, review.EditReview(reviewData, movieId))
}

func AddMovieReview(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId := c.Query("movieId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return utils.Render(c, review.AddReview(movieId))
}

func InsertMovieReview(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId := c.QueryInt("movieId")
	userId := c.Locals("UserId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		IsPrivate bool   `form:"review_private"`
		Review    string `form:"review"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	id := 0

	err := db.Client.Get(&id, `
INSERT INTO review (content, private, user_id, movie_id)
    VALUES ($1, $2, $3, $4)
RETURNING
    id
		`, data.Review, data.IsPrivate, userId, movieId)

	if err != nil {
		return err
	}

	var reviewData types.Review

	err = db.Client.Get(&reviewData, `
SELECT
    id,
    content,
    private
FROM
    review
WHERE
    id = $1
		`, id)

	if err != nil {
		return err
	}

	return utils.Render(c, review.Review(reviewData, movieId))
}

func DeleteMovieReview(c *fiber.Ctx) error {
	id := c.Params("id")
	isAuth := utils.IsAuthenticated(c)
	movieId := c.QueryInt("movieId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	_, err := db.Client.Exec(`
DELETE FROM review
WHERE id = $1
		`, id)

	if err != nil {
		return err
	}

	return utils.Render(c, review.Review(types.Review{}, movieId))
}

func UpdateMovieReview(c *fiber.Ctx) error {
	q, err := db.MakeMovieQueries(c)
	movieId := c.QueryInt("movieId")

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

	return utils.Render(c, review.ReviewContent(reviewData, movieId))
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
	api := api.New(c)
	movie, err := tmdbApi.Movie()

	if err != nil {
		return err
	}

	tx := db.Client.MustBegin()

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

	slog.Debug("Movie updated")

	api.AddLanguages(tx, id, movie)
	api.AddGenres(tx, id, movie)
	api.AddCountries(tx, id, movie)
	api.AddProductionCompanies(tx, id, movie)
	api.AddCast(tx, movie.ImdbId, id)

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

	if err != nil {
		return err
	}

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
