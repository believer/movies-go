package router

import (
	h "believer/movies/handlers"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func redirectToHome(c *fiber.Ctx) error {
	return c.Redirect("/")
}

func SetupRoutes(app *fiber.App) {
	app.Get("/health", h.GetHealth)
	app.Get("/", h.GetFeed)

	// Auth
	// --------------------------
	app.Get("/login", h.GetLogin)
	app.Post("/login", h.Login)
	app.Post("/logout", h.Logout)

	// Watchlist
	// --------------------------
	watchlistGroup := app.Group("/watchlist")

	watchlistGroup.Get("/", h.GetWatchlist)
	watchlistGroup.Get("/movies", h.GetWatchlistMovies)
	watchlistGroup.Get("/unreleased-movies", h.GetWatchlistUnreleasedMovies)
	watchlistGroup.Delete("/:id", h.DeleteFromWatchlist)

	// Movies
	// --------------------------
	movieGroup := app.Group("/movie")

	movieGroup.Get("/", redirectToHome)
	movieGroup.Get("/imdb", h.GetByImdbId)
	movieGroup.Get("/search", h.HandleSearch)
	movieGroup.Get("/new", h.GetMovieNew)
	movieGroup.Get("/new/series", h.GetMovieNewSeries)
	// TODO: Should really be post to /movie
	movieGroup.Post("/new", h.PostMovieNew)

	movieGroup.Get("/:id", h.GetMovieByID)
	movieGroup.Patch("/:id", h.UpdateMovieByID)
	movieGroup.Get("/:imdbId/awards", h.GetMovieAwards)
	movieGroup.Get("/:id/seen/others", h.GetMovieOthersSeenByID)
	movieGroup.Post("/:id/seen", h.CreateSeenMovie)
	movieGroup.Delete("/:id/seen/:seenId", h.DeleteSeenMovie)
	movieGroup.Get("/:id/seen/:seenId/edit", h.GetSeenMovie)
	movieGroup.Put("/:id/seen/:seenId/edit", h.UpdateSeenMovie)

	movieGroup.Get("/:id/rating", h.GetRating)
	movieGroup.Get("/:id/rating/edit", h.GetEditRating)
	movieGroup.Post("/:id/rating", h.PostRating)
	movieGroup.Put("/:id/rating", h.UpdateRating)
	movieGroup.Delete("/:id/rating", h.DeleteRating)

	movieGroup.Get("/:id/watch-providers", h.WatchProviders)

	// Review
	// --------------------------
	reviewGroup := app.Group("/review")

	reviewGroup.Get("/new", h.AddMovieReview)
	reviewGroup.Post("/new", h.InsertMovieReview)
	reviewGroup.Get("/:id/edit", h.EditMovieReview)
	reviewGroup.Put("/:id", h.UpdateMovieReview)
	reviewGroup.Delete("/:id", h.DeleteMovieReview)

	// Year
	// --------------------------
	yearGroup := app.Group("/year")

	yearGroup.Get("/", redirectToHome)
	yearGroup.Get("/:year", h.GetMoviesByYear)

	// Person
	// --------------------------
	personGroup := app.Group("/person")

	personGroup.Get("/", redirectToHome)
	personGroup.Get("/:id", h.GetPersonByID)

	// Genre
	// --------------------------
	genreGroup := app.Group("/genre")

	genreGroup.Get("/", redirectToHome)
	genreGroup.Get("/stats", h.GetGenreStats)
	genreGroup.Get("/:id", h.GetGenre)

	// Language
	// --------------------------
	languageGroup := app.Group("/language")

	languageGroup.Get("/", redirectToHome)
	languageGroup.Get("/stats", h.GetLanguageStats)
	languageGroup.Get("/:id", h.GetLanguage)

	// Production companies
	// --------------------------
	productionCompanyGroup := app.Group("/production-company")

	productionCompanyGroup.Get("/", redirectToHome)
	productionCompanyGroup.Get("/stats", h.GetProductionCompanyStats)
	productionCompanyGroup.Get("/:id", h.GetProductionCompany)

	// Production countries
	// --------------------------
	productionCountryGroup := app.Group("/production-country")

	productionCountryGroup.Get("/", redirectToHome)
	productionCountryGroup.Get("/stats", h.GetProductionCountryStats)
	productionCountryGroup.Get("/:id", h.GetProductionCountry)

	// Awards
	// --------------------------
	awardsGroup := app.Group("/awards")

	awardsGroup.Get("/", redirectToHome)
	awardsGroup.Get("/:awards", h.GetMoviesByNumberOfAwards)
	awardsGroup.Get("/year/:year", h.GetAwardsByYear)

	// Rating
	// --------------------------
	ratingGroup := app.Group("/rating")

	ratingGroup.Get("/", redirectToHome)
	ratingGroup.Get("/:rating", h.GetMoviesByRating)

	// Stats
	// --------------------------
	statsGroup := app.Group("/stats")

	statsGroup.Get("/", h.GetStats)
	statsGroup.Get("/ratings", h.GetRatingsByYear)
	statsGroup.Get("/ratings/:year", h.GetRatingsForYear)
	statsGroup.Get("/by-month", h.GetThisYearByMonth)
	statsGroup.Get("/highest-ranked-person", h.GetHighestRankedPersonByJob)
	statsGroup.Get("/best-of-the-year", h.GetBestOfTheYear)
	statsGroup.Get("/most-watched-person/:job", h.GetMostWatchedByJob)

	// Series
	// --------------------------
	seriesGroup := app.Group("/series")

	seriesGroup.Get("/:id", h.GetSeries)

	// Settings
	// --------------------------
	settingsGroup := app.Group("/settings")

	settingsGroup.Get("/", h.Settings)
	settingsGroup.Put("/watch-providers", h.SettingsWatchProviders)

	hookGroup := app.Group("/hooks")

	hookGroup.Post("/progress", func(c *fiber.Ctx) error {
		fmt.Println(string(c.Body()))
		return c.SendStatus(fiber.StatusOK)
	})

	hookGroup.Post("/stopped", func(c *fiber.Ctx) error {
		fmt.Println(string(c.Body()))
		return c.SendStatus(fiber.StatusOK)
	})
}
