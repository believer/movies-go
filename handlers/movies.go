package handlers

import (
	"believer/movies/components/list"
	"believer/movies/components/movie"
	"believer/movies/components/rating"
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

type MovieHandler struct {
	repo db.MovieQuerier
}

func NewMovieHandler(repo db.MovieQuerier) *MovieHandler {
	return &MovieHandler{repo}
}

func (h *MovieHandler) GetMovieByID(c *fiber.Ctx) error {
	backParam := c.QueryBool("back", false)
	movieId := c.Params("id")
	id, err := utils.SelfHealingUrl(movieId)
	if err != nil {
		id = "0"
	}
	userID := c.Locals("UserId").(string)

	movieData, err := h.repo.GetByID(id, userID)

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

	reviewData, err := h.repo.GetReviewByMovieID(id, userID)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	isInWatchlist, err := h.repo.IsWatchlisted(id, userID)

	if err != nil {
		return err
	}

	others, err := h.repo.RatingsByOthers(id)

	if err != nil {
		return err
	}

	watchedAt, err := h.repo.SeenByUser(id, userID)

	if err != nil {
		return err
	}

	cast, hasCharacters, err := h.repo.Cast(id)

	if err != nil {
		return err
	}

	if c.Get("Accept") == "application/json" {
		return c.JSON(movieData)
	}

	return utils.Render(c, views.Movie(
		views.MovieProps{
			Cast:          cast,
			HasCharacters: hasCharacters,
			WatchedAt:     watchedAt,
			IsInWatchlist: isInWatchlist,
			Movie:         movieData,
			Others:        others,
			Review:        reviewData,
			Back:          backParam,
		}))
}

func (h *MovieHandler) GetMovieOthersSeenByID(c *fiber.Ctx) error {
	movieId := c.Params("id")
	id, err := utils.SelfHealingUrl(movieId)
	if err != nil {
		id = "0"
	}

	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		return err
	}

	others, err := h.repo.RatingsByOthers(id)

	if err != nil {
		return err
	}

	return utils.Render(c, seen.MovieOthersSeen(seen.MovieOthersSeenProps{
		ID:     idAsInt,
		Others: others,
	}))
}

// Render the add movie page
func (h *MovieHandler) GetMovieNew(c *fiber.Ctx) error {
	var movieData types.Movie

	isAuth := utils.IsAuthenticated(c)
	id := c.QueryInt("id")
	imdbId := c.Query("imdbId")
	userID, ok := c.Locals("UserId").(string)
	if !ok {
		userID = ""
	}

	if !isAuth {
		return c.Redirect("/")
	}

	friends, err := h.repo.GetFriends(userID)

	if err != nil {
		return err
	}

	if id == 0 {
		return utils.Render(c, views.NewMovie(views.NewMovieProps{
			Friends:     friends,
			ImdbID:      imdbId,
			InWatchlist: false,
			Movie:       movieData,
		}))
	}

	isInWatchlist, err := h.repo.IsWatchlisted(strconv.Itoa(id), userID)

	if err != nil {
		return err
	}

	movieData, err = h.repo.GetMovieByIDSimple(id)

	if err != nil {
		return err
	}

	return utils.Render(c, views.NewMovie(views.NewMovieProps{
		Friends:     friends,
		ImdbID:      imdbId,
		InWatchlist: isInWatchlist,
		Movie:       movieData,
	}))
}

func (h *MovieHandler) GetMovieNewSeries(c *fiber.Ctx) error {
	options, err := h.repo.GetAllSeries()

	if err != nil {
		return err
	}

	return utils.Render(c, list.DataList(options, "series_list"))
}

// Handle adding a movie
func (h *MovieHandler) PostMovieNew(c *fiber.Ctx) error {
	if c.Locals("IsAuthenticated") == false {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		HasWilhelmScream bool     `form:"wilhelm_scream"`
		ImdbID           string   `form:"imdb_id"`
		IsPrivateReview  bool     `form:"review_private"`
		IsWatchlist      bool     `form:"watchlist"`
		NumberInSeries   int      `form:"number_in_series"`
		Rating           int      `form:"rating"`
		Review           string   `form:"review"`
		Series           string   `form:"series"`
		Friend           []string `form:"friend"`
		WatchedAt        string   `form:"watched_at"`
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

	movieData, movieId, err := api.AddMovie(imdbId, data.HasWilhelmScream)
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
	defer tx.Rollback()

	// Add review if any
	if data.Review != "" {
		err = h.repo.InsertReview(tx, data.Review, data.IsPrivateReview, userId, movieId)
		if err != nil {
			return err
		}
	}

	// Insert series
	if data.Series != "" && data.NumberInSeries != 0 {
		seriesId, err := strconv.Atoi(data.Series)

		// Series can't be turned into an int, so it's a new series
		if err != nil {
			seriesId, err = h.repo.GetOrInsertSeries(tx, data.Series)
			if err != nil {
				return err
			}
		}

		err = h.repo.InsertMovieSeries(tx, movieId, seriesId, data.NumberInSeries)

		if err != nil {
			c.Set("HX-Retarget", "#error")
			return c.SendString(fmt.Sprintf("Movie #%d already exists in series", data.NumberInSeries))
		}

		slog.Debug("Series inserted")
	}

	if data.IsWatchlist {
		// Add to watchlist
		err = h.repo.InsertWatchlist(tx, userId, movieId)

		if err != nil {
			c.Set("HX-Retarget", "#error")
			return c.SendString("Movie already added to watchlist")
		}
	} else {
		// Insert a view and delete from watchlist if exists
		seenID, err := h.repo.InsertSeenMovie(tx, userId, movieId, watchedAt)

		if err != nil {
			return err
		}

		err = h.repo.DeleteWatchlist(tx, userId, movieId)
		if err != nil {
			return err
		}

		// Add friends if any
		if seenID != 0 && len(data.Friend) > 0 {
			err = h.repo.InsertSeenWith(tx, seenID, data.Friend)

			if err != nil {
				return err
			}
		}
	}

	// Remove from now playing if exists
	err = h.repo.DeleteNowPlaying(tx, userId, movieData.ImdbId)
	if err != nil {
		return err
	}

	// Insert rating
	if data.Rating != 0 {
		err = h.repo.AddRating(tx, userId, movieId, data.Rating)
		if err != nil {
			return err
		}
	}

	// Add awards
	awards.AddOscars(tx, movieData.ImdbId)
	awards.AddBaftas(tx, movieData.ImdbId)

	err = tx.Commit()
	if err != nil {
		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%d", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func (h *MovieHandler) GetByImdbId(c *fiber.Ctx) error {
	imdbId := c.Query("imdbId")

	movieData, err := h.repo.GetMovieByImdbID(imdbId)

	if err != nil || movieData.ID == 0 {
		return c.SendString("")
	}

	return utils.Render(c, views.MovieExists(movieData))
}

func (h *MovieHandler) DeleteSeenMovie(c *fiber.Ctx) error {
	movieId, err := c.ParamsInt("id")
	seenId := c.Params("seenId")
	isAuth := utils.IsAuthenticated(c)
	userID := c.Locals("UserId").(string)

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tx := db.Client.MustBegin()
	defer tx.Rollback()

	err = h.repo.DeleteSeenMovie(tx, seenId)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	watchedAt, err := h.repo.SeenByUser(strconv.Itoa(movieId), userID)

	if err != nil {
		return err
	}

	return utils.Render(c, movie.Watched(movie.WatchedProps{
		WatchedAt: watchedAt,
		ID:        movieId,
	}))
}

func (h *MovieHandler) GetSeenMovie(c *fiber.Ctx) error {
	movieId, err := c.ParamsInt("id")
	seenId := c.Params("seenId")
	userId := c.Locals("UserId").(string)

	if err != nil {
		return err
	}

	watch, err := h.repo.GetSeenMovie(seenId)

	if err != nil {
		return err
	}

	friends, err := h.repo.GetFriends(userId)

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

func (h *MovieHandler) UpdateSeenMovie(c *fiber.Ctx) error {
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

	err = h.repo.UpdateSeenMovie(seenID, watchedAt, data.Friend)

	if err != nil {
		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%d", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func (h *MovieHandler) CreateSeenMovie(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	userID := c.Locals("UserId").(string)
	movieId := c.Params("id")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	err := h.repo.CreateSeenMovieDirect(userID, movieId)

	if err != nil {
		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%s?back=true", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func (h *MovieHandler) HandleSearch(c *fiber.Ctx) error {
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

func (h *MovieHandler) DeleteRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")
	userId := c.Locals("UserId").(string)

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tx := db.Client.MustBegin()
	defer tx.Rollback()

	err = h.repo.DeleteRating(tx, movieId, userId)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return utils.Render(c, rating.AddRating(rating.AddRatingProps{
		MovieId: movieId,
	}))
}

func (h *MovieHandler) GetRating(c *fiber.Ctx) error {
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

func (h *MovieHandler) GetEditRating(c *fiber.Ctx) error {
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

func (h *MovieHandler) PostRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")
	userId := c.Locals("UserId").(string)

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

	ratingVal, err := strconv.Atoi(data.Rating)
	if err != nil {
		return err
	}

	tx := db.Client.MustBegin()
	defer tx.Rollback()

	err = h.repo.AddRating(tx, userId, movieId, ratingVal)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	movieRating := int64(ratingVal)

	return utils.Render(c, rating.GetRating(rating.Props{
		MovieId: movieId,
		Rating:  movieRating,
		RatedAt: time.Now(),
	}))
}

func (h *MovieHandler) UpdateRating(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId, err := c.ParamsInt("id")
	userId := c.Locals("UserId").(string)

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

	err = h.repo.UpdateRating(userId, movieId, data.Rating)

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

func (h *MovieHandler) GetMovieAwards(c *fiber.Ctx) error {
	var year string

	imdbId := c.Params("imdbId")
	awardType := c.Query("type")

	awardsList, err := h.repo.GetMovieAwards(imdbId, awardType)

	if err != nil {
		return err
	}

	won := 0

	for _, award := range awardsList {
		year = award.Year

		if award.Winner {
			won++
		}
	}

	return utils.Render(c, views.MovieAwards(views.MovieAwardsProps{
		Awards: awardsList,
		Type:   awardType,
		Year:   year,
		Won:    won,
	}))
}

func (h *MovieHandler) UpdateMovieByID(c *fiber.Ctx) error {
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

	movieSimple, err := h.repo.GetMovieTitleAndImdbID(strconv.Itoa(id))

	if err != nil {
		return err
	}

	tmdbApi := tmdb.New(movieSimple.ImdbId)
	api := api.New(c)
	movieData, err := tmdbApi.Movie()

	if err != nil {
		return err
	}

	tx := db.Client.MustBegin()
	defer tx.Rollback()

	err = h.repo.UpdateMovie(tx, id, movieData.Title, movieData.Runtime, movieData.ReleaseDate, movieData.ImdbId, movieData.Overview, movieData.Poster, movieData.Tagline, movieData.TmdbId)
	if err != nil {
		return err
	}

	slog.Debug("Movie updated")

	api.AddLanguages(tx, id, movieData)
	api.AddGenres(tx, id, movieData)
	api.AddCountries(tx, id, movieData)
	api.AddProductionCompanies(tx, id, movieData)
	api.AddCast(tx, movieData.ImdbId, id)

	// Add awards
	awards.AddOscars(tx, movieData.ImdbId)
	awards.AddBaftas(tx, movieData.ImdbId)

	err = tx.Commit()
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *MovieHandler) WatchProviders(c *fiber.Ctx) error {
	movieId := c.Params("id")
	id, err := utils.SelfHealingUrl(movieId)
	if err != nil {
		id = "0"
	}
	userID := c.Locals("UserId").(string)

	storedProviders, err := h.repo.GetUserWatchProviders(userID)

	if err != nil {
		return err
	}

	movieObj, err := h.repo.GetMovieTitleAndImdbID(id)

	if err != nil {
		return err
	}

	t := tmdb.New(movieObj.ImdbId)
	watchProviders, err := t.WatchProviders()

	if err != nil {
		return err
	}

	hasProviders := len(watchProviders.Results.SE.Buy) > 0 || len(watchProviders.Results.SE.Rent) > 0 || len(watchProviders.Results.SE.Subscription) > 0

	if !hasProviders {
		c.Status(fiber.StatusNotFound)
		return c.SendString("")
	}

	// Only Swedish providers supported
	justWatchUrl := fmt.Sprintf("https://www.justwatch.com/se/film/%s", utils.Slugify(movieObj.Title))

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
		HasMyProviders:         hasMyProviders,
		HasOtherProviders:      hasOtherProviders,
		MissingStoredProviders: storedProviders == "",
		OtherProviders:         otherProviders,
		MyProviders:            myProviders,
		JustWatchLink:          justWatchUrl,
	}))
}

type Progress struct {
	Completed bool   `form:"completed"`
	ImdbID    string `form:"imdb_id"`
	Name      string `json:"name"`
	Position  string `json:"position"`
}

func (h *MovieHandler) PlaybackProgress(c *fiber.Ctx) error {
	var data Progress

	if c.Locals("IsAuthenticated") == false {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	if data.ImdbID == "" {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if data.Completed {
		slog.Info("Playback completed", "data", data)
		return h.PostMovieNew(c)
	} else {
		// Convert string position to float
		positionParts := strings.Split(data.Position, ":")
		positionAsNumber := 0.0

		for i, p := range positionParts {
			n, err := strconv.Atoi(p)

			if err != nil {
				continue
			}

			switch i {
			case 0:
				positionAsNumber += 60 * float64(n)
			case 1:
				positionAsNumber += float64(n)
			case 2:
				positionAsNumber += float64(n) / 60
			}
		}

		userID := c.Locals("UserId").(string)
		err := h.repo.UpdateNowPlaying(data.ImdbID, positionAsNumber, userID)

		if err != nil {
			return err
		}

		// If movie doesn't exist, add it
		movieExists, err := h.repo.MovieExists(data.ImdbID)

		if err != nil {
			return err
		}

		if movieExists {
			// Try to update existing movie
			_ = h.UpdateMovieByID(c)
		} else {
			api := api.New(c)
			_, _, err := api.AddMovie(data.ImdbID, false)

			if err != nil {
				return err
			}
		}

		slog.Info("Playback updated", "data", data)
	}

	return c.SendStatus(fiber.StatusOK)
}

