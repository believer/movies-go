package router

import (
	"believer/movies/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/health", handlers.HandleHealthCheck)
	app.Get("/", handlers.HandleFeed)

	// Movies
	// --------------------------
	movieGroup := app.Group("/movies")
	movieGroup.Get("/:id", handlers.HandleGetMovieByID)
	movieGroup.Get("/:id/cast", handlers.HandleGetMovieCastByID)
	movieGroup.Get("/:id/seen", handlers.HandleGetMovieSeenByID)

	// Person
	// --------------------------
	personGroup := app.Group("/person")
	personGroup.Get("/:id", handlers.HandleGetPersonByID)

	// Search
	// --------------------------
	app.Post("/search", handlers.HandleMovieSearch)
}
