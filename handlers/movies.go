package handlers

import (
	"believer/movies/components/list"
	"believer/movies/components/movie"
	"believer/movies/components/rating"
	"believer/movies/components/seen"
	"believer/movies/db"
	"believer/movies/services/intake"
	"believer/movies/services/tmdb"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type MovieHandler struct {
	repo   db.MovieQuerier
	intake *intake.Intake
}

func NewMovieHandler(repo db.MovieQuerier, intake *intake.Intake) *MovieHandler {
	return &MovieHandler{repo, intake}
}

func (h *MovieHandler) GetMovieByID(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	backParam := req.QueryBool("back", false)
	id := req.MovieID()
	userID := req.UserID()

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
		return c.JSON(map[string]any{
			"movie":         movieData,
			"cast":          cast,
			"review":        reviewData,
			"othersRating":  others,
			"isWatchlisted": isInWatchlist,
			"watchedAt":     watchedAt,
		})
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
	req := utils.NewRequest(c)
	id := req.MovieID()

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
	req := utils.NewRequest(c)
	var movieData types.Movie

	isAuth := req.IsAuthenticated()
	id := req.QueryInt("id")
	imdbId := req.Query("imdbId")
	userID := req.UserID()

	if !isAuth {
		return c.Redirect("/")
	}

	friends, err := h.repo.GetFriends(userID)

	if err != nil {
		return err
	}

	if id == 0 {
		if req.IsJson() {
			return c.JSON(map[string]any{
				"friends":       friends,
				"isWatchlisted": false,
				"movie":         movieData,
			})
		}

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

	if req.IsJson() {
		return c.JSON(map[string]any{
			"friends":       friends,
			"isWatchlisted": isInWatchlist,
			"movie":         movieData,
		})
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
	req := utils.NewRequest(c)
	if !req.IsAuthenticated() {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		DryRun           bool     `form:"dry_run"`
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
	userId := req.UserID()

	if err != nil {
		c.Set("HX-Retarget", "#error")
		return c.SendString(err.Error())
	}

	opts := intake.MovieOptions{WilhelmScream: data.HasWilhelmScream}

	if data.Series != "" && data.NumberInSeries != 0 {
		series := intake.SeriesEntry{Number: float64(data.NumberInSeries)}

		// Series can't be turned into an int, so it's a new series
		if seriesId, err := strconv.Atoi(data.Series); err == nil {
			series.ID = seriesId
		} else {
			series.Name = data.Series
		}

		opts.Series = &series
	}

	ref := intake.ByExternalID(imdbId)
	var movieId int

	switch {
	case data.DryRun:
		movieId, err = h.intake.EnsureMovie(ref, opts)
	case data.IsWatchlist:
		movieId, err = h.intake.AddToWatchlist(userId, ref, opts)
	default:
		watch := intake.Watch{
			WatchedAt: utils.ParseWatchedAt(data.WatchedAt, time.Now()),
			Friends:   data.Friend,
			Rating:    data.Rating,
		}

		if data.Review != "" {
			watch.Review = &intake.Review{Content: data.Review, Private: data.IsPrivateReview}
		}

		movieId, err = h.intake.LogWatch(userId, ref, opts, watch)
	}

	switch {
	case errors.Is(err, intake.ErrAlreadyInSeries):
		c.Set("HX-Retarget", "#error")
		return c.SendString(fmt.Sprintf("Movie #%d already exists in series", data.NumberInSeries))
	case errors.Is(err, intake.ErrAlreadyWatchlisted):
		c.Set("HX-Retarget", "#error")
		return c.SendString("Movie already added to watchlist")
	case errors.Is(err, intake.ErrMovieNotFound):
		c.Set("HX-Retarget", "#error")
		return c.SendString("Movie not found on TMDB")
	case err != nil:
		return err
	}

	InvalidateStatsCache(userId)

	if c.Get("Accept") == "text/plain" {
		return c.SendString(strconv.Itoa(movieId))
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%d", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func (h *MovieHandler) GetByImdbId(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	imdbId := req.Query("imdbId")

	movieData, err := h.repo.GetMovieByImdbID(imdbId)

	if err != nil || movieData.ID == 0 {
		return c.SendString("")
	}

	return utils.Render(c, views.MovieExists(movieData))
}

func (h *MovieHandler) DeleteSeenMovie(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	movieId, err := req.ParamsInt("id")
	seenId := req.Params("seenId")
	isAuth := req.IsAuthenticated()
	userID := req.UserID()

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tx, err := h.repo.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = h.repo.DeleteSeenMovie(tx, seenId)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	InvalidateStatsCache(userID)

	watchedAt, err := h.repo.SeenByUser(strconv.Itoa(movieId), userID)

	if err != nil {
		return err
	}

	return utils.Render(c, movie.Watched(movie.WatchedProps{
		WatchedAt: watchedAt,
		ID:        movieId,
	}))
}

func (h *MovieHandler) DeleteNowPlaying(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	movieIdStr := req.Params("id")
	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		return err
	}
	isAuth := req.IsAuthenticated()
	userID := req.UserID()

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	movieSimple, err := h.repo.GetMovieTitleAndImdbID(movieIdStr)
	if err != nil {
		return err
	}

	err = h.repo.DeleteNowPlayingDirect(userID, movieId)
	if err != nil {
		return err
	}

	InvalidateStatsCache(userID)

	watchedAt, err := h.repo.SeenByUser(movieIdStr, userID)
	if err != nil {
		return err
	}

	isInWatchlist, err := h.repo.IsWatchlisted(movieIdStr, userID)
	if err != nil {
		return err
	}

	return utils.Render(c, movie.Watched(movie.WatchedProps{
		WatchedAt:   watchedAt,
		IsUnseen:    len(watchedAt) == 0,
		InWatchlist: isInWatchlist,
		ImdbId:      movieSimple.ImdbId,
		ID:          movieId,
		NowPlaying:  nil,
	}))
}

func (h *MovieHandler) GetSeenMovie(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	movieId, err := req.ParamsInt("id")
	seenId := req.Params("seenId")
	userId := req.UserID()

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
	req := utils.NewRequest(c)
	movieId, err := req.ParamsInt("id")
	seenID := req.Params("seenId")
	isAuth := req.IsAuthenticated()

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

	watchedAt := utils.ParseWatchedAt(data.WatchedAt, time.Now())

	err = h.repo.UpdateSeenMovie(seenID, watchedAt, data.Friend)

	if err != nil {
		return err
	}

	if userID := req.UserID(); userID != "" {
		InvalidateStatsCache(userID)
	}

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%d", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func (h *MovieHandler) CreateSeenMovie(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	userID := req.UserID()
	movieId := req.Params("id")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	id, err := strconv.Atoi(movieId)

	if err != nil {
		return err
	}

	_, err = h.intake.LogWatch(userID, intake.ByID(id), intake.MovieOptions{}, intake.Watch{WatchedAt: time.Now()})

	if errors.Is(err, intake.ErrMovieNotFound) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if err != nil {
		return err
	}

	InvalidateStatsCache(userID)

	c.Set("HX-Redirect", fmt.Sprintf("/movie/%s?back=true", movieId))

	return c.SendStatus(fiber.StatusOK)
}

func (h *MovieHandler) HandleSearch(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	query := req.Query("search")

	if query == "" {
		return utils.Render(c, views.MovieSearch([]types.SearchResult{}))
	}

	tmdbApi := tmdb.New("")
	movies, err := tmdbApi.Search(query)

	if err != nil {
		return err
	}

	if req.IsJson() {
		return c.JSON(movies)
	}

	if len(movies.Results) == 0 {
		return utils.Render(c, views.MovieSearchEmpty())
	}

	maxResults := clamp(len(movies.Results), 1, 5)

	return utils.Render(c, views.MovieSearch(movies.Results[:maxResults]))
}

func (h *MovieHandler) DeleteRating(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	movieId, err := req.ParamsInt("id")
	userId := req.UserID()

	if err != nil {
		return err
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tx, err := h.repo.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = h.repo.DeleteRating(tx, movieId, userId)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	InvalidateStatsCache(userId)

	return utils.Render(c, rating.AddRating(rating.AddRatingProps{
		MovieId: movieId,
	}))
}

func (h *MovieHandler) GetRating(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	movieId, err := req.ParamsInt("id")
	currentRating := req.QueryInt("rating")

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
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	movieId, err := req.ParamsInt("id")

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
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	movieId, err := req.ParamsInt("id")
	userId := req.UserID()

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

	tx, err := h.repo.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = h.repo.AddRating(tx, userId, movieId, ratingVal)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	InvalidateStatsCache(userId)

	movieRating := int64(ratingVal)

	return utils.Render(c, rating.GetRating(rating.Props{
		MovieId: movieId,
		Rating:  movieRating,
		RatedAt: time.Now(),
	}))
}

func (h *MovieHandler) UpdateRating(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	movieId, err := req.ParamsInt("id")
	userId := req.UserID()

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

	InvalidateStatsCache(userId)

	movieRating, _ := strconv.ParseInt(data.Rating, 10, 0)

	return utils.Render(c, rating.GetRating(rating.Props{
		MovieId: movieId,
		Rating:  movieRating,
		RatedAt: time.Now(),
	}))
}

func (h *MovieHandler) GetMovieAwards(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	var year string

	imdbId := req.Params("imdbId")
	awardType := req.Query("type")

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
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userId := req.UserID()

	if userId != "1" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	id, err := req.ParamsInt("id")

	if err != nil {
		return err
	}

	err = h.intake.RefreshMovie(id)

	if errors.Is(err, intake.ErrMovieNotFound) {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if err != nil {
		return err
	}

	InvalidateStatsCache(userId)

	return c.SendStatus(fiber.StatusOK)
}

func (h *MovieHandler) WatchProviders(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	id := req.MovieID()
	userID := req.UserID()

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

	imdbID, err := utils.ParseId(data.ImdbID)

	if err != nil {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	userID := c.Locals("UserId").(string)
	ref := intake.ByExternalID(imdbID)

	if data.Completed {
		_, err := h.intake.LogWatch(userID, ref, intake.MovieOptions{}, intake.Watch{WatchedAt: time.Now()})

		if err != nil {
			return err
		}

		InvalidateStatsCache(userID)
		slog.Info("Playback completed", "data", data)
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

		movieId, err := h.intake.EnsureMovie(ref, intake.MovieOptions{})

		if err != nil {
			return err
		}

		err = h.repo.UpdateNowPlaying(movieId, positionAsNumber, userID)

		if err != nil {
			return err
		}

		slog.Info("Playback updated", "data", data)
	}

	return c.SendStatus(fiber.StatusOK)
}
