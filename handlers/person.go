package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"
	"sort"

	"github.com/gofiber/fiber/v2"
)

func GetPersonByID(c *fiber.Ctx) error {
	var person types.Person
	var awards []types.Award

	q := db.MakeQueries(c)

	err := db.Client.Get(&person, `
SELECT
    p.id,
    p.name,
    -- Function get_person_role_json returns a JSON array of movies
    -- The function is defined in the database
    get_person_role_with_seen_json (p.id, 'director'::job, $2) AS director,
    get_person_role_with_seen_json (p.id, 'cast', $2) AS cast,
    get_person_role_with_seen_json (p.id, 'writer', $2) AS writer,
    get_person_role_with_seen_json (p.id, 'composer', $2) AS composer,
    get_person_role_with_seen_json (p.id, 'producer', $2) AS producer,
    get_person_role_with_seen_json (p.id, 'cinematographer', $2) AS cinematographer,
    get_person_role_with_seen_json (p.id, 'editor', $2) AS editor
FROM
    person AS p
WHERE
    p.id = $1
		`, q.Id, q.UserId)

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Person not found")
		}

		return err
	}

	err = db.Client.Select(&awards, `
SELECT
    a.name AS "category",
    a.detail,
    a.winner,
    a.year,
    m.title,
    m.id AS "movie_id"
FROM
    award a
    INNER JOIN movie m ON m.imdb_id = a.imdb_id
WHERE
    person_id = $1
ORDER BY
    a.year DESC
		`, q.Id)

	if err != nil {
		return err
	}

	fields := []int{
		len(person.Cast),
		len(person.Director),
		len(person.Writer),
		len(person.Producer),
		len(person.Composer),
		len(person.Cinematographer),
		len(person.Editor),
	}

	totalCredits := 0
	for _, field := range fields {
		totalCredits += field
	}

	groupedAwards := make(map[string][]types.Award)

	won := 0
	for _, award := range awards {
		if award.Winner {
			won++
		}

		groupedAwards[award.Category] = append(groupedAwards[award.Category], award)
	}

	// Awards map is unsorted, create a sort order
	awardsOrder := make([]string, 0, len(groupedAwards))
	for k := range groupedAwards {
		awardsOrder = append(awardsOrder, k)
	}

	sort.Strings(awardsOrder)

	return utils.Render(c, views.Person(views.PersonProps{
		Awards:       groupedAwards,
		AwardsOrder:  awardsOrder,
		Person:       person,
		TotalCredits: totalCredits,
		Won:          won,
	}))
}
