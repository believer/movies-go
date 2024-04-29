package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func HandleGetPersonByID(c *fiber.Ctx) error {
	var person types.Person

	id := c.Params("id")

	err := db.Dot.Get(db.Client, &person, "person-by-id", id)

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Person not found")
		}

		return err
	}

	fields := []int{
		len(person.Cast),
		len(person.Director),
		len(person.Writer),
		len(person.Producer),
		len(person.Composer),
	}

	totalCredits := 0
	for _, field := range fields {
		totalCredits += field
	}

	return utils.TemplRender(c, views.Person(person, totalCredits, id))
}

func HandleSeenMovieByID(c *fiber.Ctx) error {
	var seen bool

	err := db.Dot.Get(db.Client, &seen, "person-seen-movie-by-id", c.Locals("UserId"), c.Params("movieId"))

	if err != nil {
		return err
	}

	return utils.TemplRender(c, views.Seen(seen))
}
