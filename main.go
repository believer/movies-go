package main

import (
	"believer/movies/db"
	"believer/movies/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	db.InitializeConnection()

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})

	app.Get("/", routes.FeedHandler)
	app.Get("/movies/:id", routes.MovieHandler)

	app.Static("/public", "./public")

	app.Listen(":8080")
}
