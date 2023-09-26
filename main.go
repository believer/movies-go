package main

import (
	"believer/movies/db"
	"believer/movies/routes"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	db.InitializeConnection()

	engine := html.New("./views", ".html")

	engine.AddFunc("StringsJoin", strings.Join)

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})

	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/", routes.FeedHandler)

	movieGroup := app.Group("/movies")
	movieGroup.Get("/:id", routes.MovieHandler)
	movieGroup.Get("/:id/cast", routes.MovieCastHandler)
	movieGroup.Get("/:id/seen", routes.MovieSeenHandler)

	personGroup := app.Group("/person")
	personGroup.Get("/:id", routes.PersonHandler)

	app.Post("/search", routes.SearchHandler)

	app.Static("/public", "./public")

	app.Listen(":8080")
}
