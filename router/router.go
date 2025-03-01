package router

import (
	h "believer/movies/handlers"

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
	movieGroup.Post("/new", h.PostMovieNew)

	movieGroup.Get("/:id", h.GetMovieByID)
	movieGroup.Get("/:imdbId/awards", h.GetMovieAwards)
	movieGroup.Get("/:id/cast", h.GetMovieCastByID)
	movieGroup.Get("/:id/seen", h.GetMovieSeenByID)
	movieGroup.Post("/:id/seen", h.CreateSeenMovie)
	movieGroup.Delete("/:id/seen/:seenId", h.DeleteSeenMovie)
	movieGroup.Get("/:id/seen/:seenId/edit", h.GetSeenMovie)
	movieGroup.Put("/:id/seen/:seenId/edit", h.UpdateSeenMovie)

	movieGroup.Get("/:id/rating", h.GetRating)
	movieGroup.Get("/:id/rating/edit", h.GetEditRating)
	movieGroup.Post("/:id/rating", h.PostRating)
	movieGroup.Put("/:id/rating", h.UpdateRating)
	movieGroup.Delete("/:id/rating", h.DeleteRating)

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
	genreGroup.Get("/:id", h.GetGenre)

	// Language
	// --------------------------
	languageGroup := app.Group("/language")

	languageGroup.Get("/", redirectToHome)
	languageGroup.Get("/:id", h.GetLanguage)

	// Stats
	// --------------------------
	statsGroup := app.Group("/stats")

	statsGroup.Get("/", h.GetStats)
	statsGroup.Get("/genres", h.GetGenreStats)
	statsGroup.Get("/languages", h.GetLanguageStats)
	statsGroup.Get("/ratings", h.GetRatingsByYear)
	statsGroup.Get("/by-month", h.GetThisYearByMonth)
	statsGroup.Get("/highest-ranked-person", h.GetHighestRankedPersonByJob)
	statsGroup.Get("/best-of-the-year", h.GetBestOfTheYear)
	statsGroup.Get("/most-watched-person/:job", h.GetMostWatchedByJob)

	// Series
	// --------------------------
	seriesGroup := app.Group("/series")

	seriesGroup.Get("/:id", h.GetSeries)
}
