package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func HandleGetPersonByID(c *fiber.Ctx) error {
	var person types.Person

	err := db.Client.Get(&person, `
SELECT 
  p.id,
  p.name, 
  -- Function get_person_role_json returns a JSON array of movies
  -- The function is defined in the database
  get_person_role_json(p.id, 'director'::job) as director,
  get_person_role_json(p.id, 'cast') as cast,
  get_person_role_json(p.id, 'writer') as writer,
  get_person_role_json(p.id, 'composer') as composer,
  get_person_role_json(p.id, 'producer') as producer
FROM person AS p
WHERE p.id = $1
`, c.Params("id"))

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Person not found")
		}

		return err
	}

	return c.Render("person", fiber.Map{
		"Person": person,
	})
}
