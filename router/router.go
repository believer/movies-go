package router

import (
	"believer/movies/handlers"

	"github.com/gofiber/fiber/v2"
)

func redirectToHome(c *fiber.Ctx) error {
	return c.Redirect("/")
}

func SetupRoutes(app *fiber.App) {
	app.Get("/health", handlers.HandleHealthCheck)
	app.Get("/", handlers.HandleFeed)
	app.Get("/login", handlers.HandleGetLogin)
	app.Post("/login", handlers.HandlePostLogin)
	app.Post("/logout", handlers.HandlePostLogout)

	// Watchlist
	watchlistGroup := app.Group("/watchlist")

	watchlistGroup.Get("/", handlers.HandleGetWatchlist)
	watchlistGroup.Delete("/:id", handlers.HandleDeleteFromWatchlist)

	// Movies
	// --------------------------
	movieGroup := app.Group("/movie")

	movieGroup.Get("/", redirectToHome)
	movieGroup.Get("/imdb", handlers.HandleGetByImdbId)
	movieGroup.Get("/search", handlers.HandleSearch)
	movieGroup.Get("/new", handlers.HandleGetMovieNew)
	movieGroup.Get("/new/series", handlers.HandleGetMovieNewSeries)
	movieGroup.Post("/new", handlers.HandlePostMovieNew)

	movieGroup.Get("/:id", handlers.HandleGetMovieByID)
	movieGroup.Get("/:id/cast", handlers.HandleGetMovieCastByID)
	movieGroup.Get("/:id/seen", handlers.HandleGetMovieSeenByID)
	movieGroup.Post("/:id/seen", handlers.HandlePostMovieSeen)
	movieGroup.Delete("/:id/seen/:seenId", handlers.HandleDeleteMovieSeen)
	movieGroup.Get("/:id/seen/:seenId/edit", handlers.HandleGetMovieSeen)
	movieGroup.Put("/:id/seen/:seenId/edit", handlers.HandleUpdateMovieSeen)

	movieGroup.Delete("/:id/rating", handlers.HandleDeleteRating)

	// Year
	// --------------------------
	yearGroup := app.Group("/year")

	yearGroup.Get("/", redirectToHome)
	yearGroup.Get("/:year", handlers.HandleGetMoviesByYear)

	// Person
	// --------------------------
	personGroup := app.Group("/person")

	personGroup.Get("/", redirectToHome)
	personGroup.Get("/:id", handlers.HandleGetPersonByID)

	// Genre
	// --------------------------
	genreGroup := app.Group("/genre")

	genreGroup.Get("/", redirectToHome)
	genreGroup.Get("/:id", handlers.HandleGetGenre)

	// Language
	// --------------------------
	languageGroup := app.Group("/language")

	languageGroup.Get("/", redirectToHome)
	languageGroup.Get("/:id", handlers.HandleGetLanguage)

	// Stats
	// --------------------------
	statsGroup := app.Group("/stats")

	statsGroup.Get("/", handlers.HandleGetStats)
	statsGroup.Get("/genres", handlers.HandleGetGenreStats)
	statsGroup.Get("/languages", handlers.HandleGetLanguageStats)
	statsGroup.Get("/ratings", handlers.HandleGetRatingsByYear)
	statsGroup.Get("/by-month", handlers.HandleGetThisYearByMonth)
	statsGroup.Get("/highest-ranked-person", handlers.HandleGetHighestRankedPersonByJob)
	statsGroup.Get("/most-watched-person/:job", handlers.HandleGetMostWatchedByJob)

	// Series
	// --------------------------
	seriesGroup := app.Group("/series")

	seriesGroup.Get("/:id", handlers.HandleGetSeries)
}
