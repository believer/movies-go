package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetMoviesByRating(c *fiber.Ctx) error {
	var movies types.Movies

	page := c.QueryInt("page", 1)
	userId := c.Locals("UserId")
	rating, err := c.ParamsInt("rating")

	if err != nil {
		return err
	}

	err = db.Client.Select(&movies, `
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    m.imdb_id,
    (s.id IS NOT NULL) AS "seen"
FROM
    rating r
    INNER JOIN movie m ON m.id = r.movie_id
    LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            public.seen
        WHERE
            user_id = $2
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id
WHERE
    r.rating = $1
    AND r.user_id = $2
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50
`, rating, userId, (page-1)*50)

	if err != nil {
		return err
	}

	// When there are no more movies to show, just return 200. Otherwise we
	// would display the "No movies seen" empty state which should only be
	// shown at the start.
	if len(movies) == 0 && page > 1 {
		return c.SendStatus(fiber.StatusOK)
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this rating",
		Name:       fmt.Sprintf("Movies rated %d", rating),
		NextPage:   fmt.Sprintf("/rating/%d?page=%d", rating, page+1),
		Movies:     movies,
	}))
}
