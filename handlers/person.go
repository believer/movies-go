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
   jsonb_agg(json_build_object('title', m.title, 'id', m.id, 'release_date', to_char(m.release_date, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))) as movies
FROM
  public.person as p
JOIN LATERAL (
   SELECT
   m.id,
      m.title,
      m.release_date
   FROM
      public.movie_person as mp
      INNER JOIN public.movie as m ON m.id = mp.movie_id
   WHERE
      p.id = mp.person_id AND mp.job = 'cast'
   ORDER BY m.release_date DESC
) as m ON true
WHERE p.id = $1
GROUP BY p.id, p.name;
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
