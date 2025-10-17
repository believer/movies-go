package db

import (
	"believer/movies/types"
	"believer/movies/utils"

	"github.com/gofiber/fiber/v2"
)

type ProductionCountry struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type ProductionCountryQueries struct {
	Id     string
	UserId string
}

func MakeProductionCountryQueries(c *fiber.Ctx) (*ProductionCountryQueries, error) {
	id := utils.SelfHealingUrlString(c.Params("id"))
	userId := c.Locals("UserId").(string)

	return &ProductionCountryQueries{
		Id:     id,
		UserId: userId,
	}, nil
}

func (pc *ProductionCountryQueries) ByID() (ProductionCountry, error) {
	var company ProductionCountry

	err := Client.Get(&company, `
		SELECT name
		FROM production_country
		WHERE id = $1
	`, pc.Id)

	if err != nil {
		return company, err
	}

	return company, nil
}

func (pc *ProductionCountryQueries) Movies(offset int) (types.Movies, error) {
	var movies types.Movies

	err := Client.Select(&movies, `
		SELECT DISTINCT
			(m.id),
			m.title,
			m.release_date,
    	(s.id IS NOT NULL) AS "seen"
		FROM movie_country mc
			INNER JOIN movie m ON m.id = mc.movie_id
			LEFT JOIN (
				SELECT DISTINCT ON (movie_id)
					movie_id, id
				FROM public.seen
				WHERE user_id = $2
				ORDER BY movie_id, id
			) AS s ON m.id = s.movie_id
		WHERE mc.country_id = $1
		ORDER BY m.release_date DESC
		OFFSET $3
		LIMIT 50
	`, pc.Id, pc.UserId, offset)

	if err != nil {
		return movies, err
	}

	return movies, nil
}
