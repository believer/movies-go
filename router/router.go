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
	app.Get("/friends", handlers.HandleGetFriends)

	// Movies
	// --------------------------
	movieGroup := app.Group("/movies")

	movieGroup.Get("/", redirectToHome)
	movieGroup.Get("/imdb", handlers.HandleGetByImdbId)
	movieGroup.Get("/new", handlers.HandleGetMovieNew)
	movieGroup.Post("/new", handlers.HandlePostMovieNew)

	movieGroup.Get("/:id", handlers.HandleGetMovieByID)
	movieGroup.Get("/:id/cast", handlers.HandleGetMovieCastByID)
	movieGroup.Get("/:id/seen", handlers.HandleGetMovieSeenByID)
	movieGroup.Post("/:id/seen", handlers.HandlePostMovieSeenNew)

	// Person
	// --------------------------
	personGroup := app.Group("/person")

	personGroup.Get("/", redirectToHome)
	personGroup.Get("/:id", handlers.HandleGetPersonByID)

	// Search
	// --------------------------
	app.Post("/search", handlers.HandleMovieSearch)

	// Stats
	// --------------------------
	statsGroup := app.Group("/stats")

	statsGroup.Get("/", handlers.HandleGetStats)
	statsGroup.Get("/most-watched-person/:job", handlers.HandleGetMostWatchedByJob)
}
