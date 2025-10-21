package db

import (
	"believer/movies/types"
	"believer/movies/utils"

	"github.com/gofiber/fiber/v2"
)

type LanguageQueries struct {
	Id     string
	UserId string
	Year   string
	Years  []string
}

func MakeLanguageQueries(c *fiber.Ctx) *LanguageQueries {
	id := utils.SelfHealingUrlString(c.Params("id"))
	userId := c.Locals("UserId").(string)
	year := c.Query("year", "All")
	years := append([]string{"All"}, utils.AvailableYears()...)

	return &LanguageQueries{
		Id:     id,
		UserId: userId,
		Year:   year,
		Years:  years,
	}
}

func (l *LanguageQueries) Stats() ([]types.ListItem, error) {
	var stats []types.ListItem

	err := Dot.Select(Client, &stats, `
SELECT
    l.id,
    l.english_name AS name,
    COUNT(DISTINCT s.movie_id) AS count
FROM ( SELECT DISTINCT ON (movie_id)
        movie_id
    FROM
        seen
    WHERE
        user_id = $1
        AND ($2 = 'All'
            OR EXTRACT(YEAR FROM date) = $2::int)) AS s
    INNER JOIN movie_language ml ON ml.movie_id = s.movie_id
    INNER JOIN "language" l ON l.id = ml.language_id
GROUP BY
    l.id
ORDER BY
    count DESC
LIMIT 10;

`, l.UserId, l.Year)

	if err != nil {
		return stats, err
	}

	return stats, nil
}
