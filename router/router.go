package router

import (
	"believer/movies/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/health", handlers.HandleHealthCheck)
	app.Get("/", handlers.HandleFeed)
	app.Get("/login", handlers.HandleGetLogin)
	app.Post("/login", handlers.HandlePostLogin)

	// Movies
	// --------------------------
	movieGroup := app.Group("/movies")

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
	personGroup.Get("/:id", handlers.HandleGetPersonByID)

	// Search
	// --------------------------
	app.Post("/search", handlers.HandleMovieSearch)
}
