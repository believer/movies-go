package router

import (
	"believer/movies/db"
	h "believer/movies/handlers"

	"github.com/gofiber/fiber/v2"
)

func redirectToHome(c *fiber.Ctx) error {
	return c.Redirect("/")
}

func SetupRoutes(app *fiber.App) {
	healthHandler := h.NewHealthHandler()
	app.Get("/health", healthHandler.GetHealth)

	feedRepo := db.NewFeedRepository(db.Client)
	nowPlayingRepo := db.NewNowPlayingRepository(db.Client)
	feedHandler := h.NewFeedHandler(feedRepo, nowPlayingRepo)
	app.Get("/", feedHandler.GetFeed)

	// Auth
	// --------------------------
	authRepo := db.NewAuthRepository(db.Client)
	authHandler := h.NewAuthHandler(authRepo)

	app.Get("/login", authHandler.GetLogin)
	app.Post("/login", authHandler.Login)
	app.Post("/logout", authHandler.Logout)

	// Watchlist
	// --------------------------
	watchlistGroup := app.Group("/watchlist")

	watchlistRepo := db.NewWatchlistRepository(db.Client)
	watchlistHandler := h.NewWatchlistHandler(watchlistRepo)

	watchlistGroup.Get("/", watchlistHandler.GetWatchlist)
	watchlistGroup.Get("/movies", watchlistHandler.GetWatchlistMovies)
	watchlistGroup.Get("/unreleased-movies", watchlistHandler.GetWatchlistUnreleasedMovies)
	watchlistGroup.Delete("/:id", watchlistHandler.DeleteFromWatchlist)

	// Movies
	// --------------------------
	movieGroup := app.Group("/movie")

	movieRepo := db.NewMovieRepository(db.Client)
	movieHandler := h.NewMovieHandler(movieRepo)

	movieGroup.Get("/", redirectToHome)
	movieGroup.Get("/imdb", movieHandler.GetByImdbId)
	movieGroup.Get("/search", movieHandler.HandleSearch)
	movieGroup.Get("/new", movieHandler.GetMovieNew)
	movieGroup.Get("/new/series", movieHandler.GetMovieNewSeries)
	// TODO: Should really be post to /movie
	movieGroup.Post("/new", movieHandler.PostMovieNew)

	movieGroup.Get("/:id", movieHandler.GetMovieByID)
	movieGroup.Patch("/:id", movieHandler.UpdateMovieByID)
	movieGroup.Get("/:imdbId/awards", movieHandler.GetMovieAwards)
	movieGroup.Get("/:id/seen/others", movieHandler.GetMovieOthersSeenByID)
	movieGroup.Post("/:id/seen", movieHandler.CreateSeenMovie)
	movieGroup.Delete("/:id/seen/:seenId", movieHandler.DeleteSeenMovie)
	movieGroup.Delete("/:id/now-playing", movieHandler.DeleteNowPlaying)
	movieGroup.Get("/:id/seen/:seenId/edit", movieHandler.GetSeenMovie)
	movieGroup.Put("/:id/seen/:seenId/edit", movieHandler.UpdateSeenMovie)

	movieGroup.Get("/:id/rating", movieHandler.GetRating)
	movieGroup.Get("/:id/rating/edit", movieHandler.GetEditRating)
	movieGroup.Post("/:id/rating", movieHandler.PostRating)
	movieGroup.Put("/:id/rating", movieHandler.UpdateRating)
	movieGroup.Delete("/:id/rating", movieHandler.DeleteRating)

	movieGroup.Get("/:id/watch-providers", movieHandler.WatchProviders)

	// Review
	// --------------------------
	reviewGroup := app.Group("/review")

	reviewRepo := db.NewReviewRepository(db.Client)
	reviewHandler := h.NewReviewHandler(reviewRepo)

	reviewGroup.Get("/new", reviewHandler.AddMovieReview)
	reviewGroup.Post("/new", reviewHandler.InsertMovieReview)
	reviewGroup.Get("/:id/edit", reviewHandler.EditMovieReview)
	reviewGroup.Put("/:id", reviewHandler.UpdateMovieReview)
	reviewGroup.Delete("/:id", reviewHandler.DeleteMovieReview)

	// Year
	// --------------------------
	yearGroup := app.Group("/year")

	yearsRepo := db.NewYearsRepository(db.Client)
	yearsHandler := h.NewYearsHandler(yearsRepo)

	yearGroup.Get("/", redirectToHome)
	yearGroup.Get("/:year", yearsHandler.GetMoviesByYear)

	// Person
	// --------------------------
	personGroup := app.Group("/person")

	personRepo := db.NewPersonRepository(db.Client)
	personHandler := h.NewPersonHandler(personRepo)

	personGroup.Get("/", redirectToHome)
	personGroup.Get("/:id", personHandler.GetPersonByID)

	// Genre
	// --------------------------
	genreGroup := app.Group("/genre")

	genreRepo := db.NewGenreRepository(db.Client)
	genreHandler := h.NewGenreHandler(genreRepo)

	genreGroup.Get("/", genreHandler.ListGenres)
	genreGroup.Get("/stats", genreHandler.GetGenreStats)
	genreGroup.Get("/:id", genreHandler.GetGenre)

	// Language
	// --------------------------
	languageGroup := app.Group("/language")

	languageRepo := db.NewLanguageRepository(db.Client)
	languageHandler := h.NewLanguageHandler(languageRepo)

	languageGroup.Get("/", languageHandler.ListLanguages)
	languageGroup.Get("/stats", languageHandler.GetLanguageStats)
	languageGroup.Get("/:id", languageHandler.GetLanguage)

	// Production companies
	// --------------------------
	productionCompanyGroup := app.Group("/production-company")

	productionCompanyRepo := db.NewProductionCompanyRepository(db.Client)
	productionCompanyHandler := h.NewProductionCompanyHandler(productionCompanyRepo)

	productionCompanyGroup.Get("/", productionCompanyHandler.ListProductionCompanies)
	productionCompanyGroup.Get("/stats", productionCompanyHandler.GetProductionCompanyStats)
	productionCompanyGroup.Get("/:id", productionCompanyHandler.GetProductionCompany)

	// Production countries
	// --------------------------
	productionCountryGroup := app.Group("/production-country")

	productionCountryRepo := db.NewProductionCountryRepository(db.Client)
	productionCountryHandler := h.NewProductionCountryHandler(productionCountryRepo)

	productionCountryGroup.Get("/", productionCountryHandler.ListProductionCountries)
	productionCountryGroup.Get("/stats", productionCountryHandler.GetProductionCountryStats)
	productionCountryGroup.Get("/:id", productionCountryHandler.GetProductionCountry)

	// Awards
	// --------------------------
	awardsGroup := app.Group("/awards")

	// Updating awards twice per year (nominations and wins)
	// - Update the CSV for award type
	// - Uncomment the route for award type
	// - Update year in awards file to only update new year
	// - After testing locally, comment out the local DATABASE_URL
	// - Run against production database

	// awardsGroup.Get("/baftas", func(c *fiber.Ctx) error {
	// 	tx := db.Client.MustBegin()
	//
	// 	awards.AddBaftas(tx, "")
	//
	// 	err := tx.Commit()
	//
	// 	if err != nil {
	// 		err = tx.Rollback()
	// 		return err
	// 	}
	//
	// 	return c.SendStatus(200)
	// })

	// awardsGroup.Get("/oscars", func(c *fiber.Ctx) error {
	// 	tx := db.Client.MustBegin()
	//
	// 	awards.AddOscars(tx, "")
	//
	// 	err := tx.Commit()
	//
	// 	if err != nil {
	// 		err = tx.Rollback()
	// 		return err
	// 	}
	//
	// 	return c.SendStatus(200)
	// })

	awardsRepo := db.NewAwardsRepository(db.Client)
	awardsHandler := h.NewAwardsHandler(awardsRepo)

	awardsGroup.Get("/", redirectToHome)
	awardsGroup.Get("/:awards", awardsHandler.GetMoviesByNumberOfAwards)
	awardsGroup.Get("/year/:year", awardsHandler.GetAwardsByYear)

	// Rating
	// --------------------------
	ratingGroup := app.Group("/rating")

	ratingsRepo := db.NewRatingsRepository(db.Client)
	ratingsHandler := h.NewRatingsHandler(ratingsRepo)

	ratingGroup.Get("/", redirectToHome)
	ratingGroup.Get("/:rating", ratingsHandler.GetMoviesByRating)

	// Stats
	// --------------------------
	statsGroup := app.Group("/stats")

	statsRepo := db.NewStatsRepository(db.Client)
	statsHandler := h.NewStatsHandler(statsRepo)

	statsGroup.Get("/", statsHandler.GetStats)
	statsGroup.Get("/ratings", statsHandler.GetRatingsByYear)
	statsGroup.Get("/wilhelm-scream", statsHandler.GetWilhelmScream)
	statsGroup.Get("/ratings/:year", statsHandler.GetRatingsForYear)
	statsGroup.Get("/by-month", statsHandler.GetThisYearByMonth)
	statsGroup.Get("/by-weekday", statsHandler.GetThisYearByWeekday)
	statsGroup.Get("/movies-by-year", statsHandler.GetMoviesByYearStat)
	statsGroup.Get("/highest-ranked-person", statsHandler.GetHighestRankedPersonByJob)
	statsGroup.Get("/best-of-the-year", statsHandler.GetBestOfTheYear)
	statsGroup.Get("/most-watched-person/:job", statsHandler.GetMostWatchedByJob)
	statsGroup.Get("/seen-with", statsHandler.GetSeenWith)

	// Series
	// --------------------------
	seriesGroup := app.Group("/series")

	seriesRepo := db.NewSeriesRepository(db.Client)
	seriesHandler := h.NewSeriesHandler(seriesRepo)

	seriesGroup.Get("/:id", seriesHandler.GetSeries)

	// Settings
	// --------------------------
	settingsGroup := app.Group("/settings")

	settingsRepo := db.NewSettingsRepository(db.Client)
	settingsHandler := h.NewSettingsHandler(settingsRepo)

	settingsGroup.Get("/", settingsHandler.Settings)
	settingsGroup.Put("/watch-providers", settingsHandler.SettingsWatchProviders)

	// Now playing
	// --------------------------
	nowPlayingGroup := app.Group("/now-playing")

	nowPlayingHandler := h.NewNowPlayingHandler(nowPlayingRepo)

	nowPlayingGroup.Get("/", nowPlayingHandler.GetNowPlaying)

	// Webhooks
	// --------------------------
	hookGroup := app.Group("/hooks")

	hookGroup.Post("/playback", movieHandler.PlaybackProgress)
}
