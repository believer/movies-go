package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func HandleGetPersonByID(c *fiber.Ctx) error {
	var person types.Person

	err := db.Dot.Get(db.Client, &person, "person-by-id", c.Params("id"))

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Person not found")
		}

		return err
	}

	if strings.Contains(c.Get("Accept"), "application/xml") {
		return c.Render("person", fiber.Map{
			"Person": person,
		})
	}

	return utils.TemplRender(c, views.Person(person))
}
