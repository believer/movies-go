package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetFeed(c *fiber.Ctx) error {
	var movies types.Movies
	var persons types.Persons

	page := c.QueryInt("page", 1)
	searchQuery := c.Query("search")
	searchQueryType := "movie"

	if searchQuery != "" {
		// Query string with a specifier for type. For example:
		// - movie:godfa
		// - actor:ryan
		if queryType, query, ok := strings.Cut(searchQuery, ":"); ok {
			switch strings.ToLower(queryType) {
			case "movie":
				err := db.Dot.Select(db.Client, &movies, "feed-search", query)

				if err != nil {
					return err
				}
			case "actor", "cast":
				err := db.Dot.Select(db.Client, &persons, "feed-search-job", query, "cast")
				searchQueryType = "person"

				if err != nil {
					return err
				}
			case "director", "writer", "producer", "composer":
				err := db.Dot.Select(db.Client, &persons, "feed-search-job", query, queryType)
				searchQueryType = "person"

				if err != nil {
					return err
				}
			}
		} else {
			err := db.Dot.Select(db.Client, &movies, "feed-search", searchQuery)

			if err != nil {
				return err
			}
		}
	} else {
		err := db.Dot.Select(db.Client, &movies, "feed", (page-1)*20, c.Locals("UserId"))

		if err != nil {
			return err
		}
	}

	// When there are no more movies to show, just return 200. Otherwise we
	// would display the "No movies seen" empty state which should only be
	// shown at the start.
	if len(movies) == 0 && page > 1 {
		return c.SendStatus(fiber.StatusOK)
	}

	if c.Get("Accept") == "application/json" {
		return c.JSON(movies)
	}

	feed := views.Feed(views.FeedProps{
		IsAdmin:   utils.IsAuthenticated(c),
		Movies:    movies,
		NextPage:  page + 1,
		Persons:   persons,
		Query:     searchQuery,
		QueryType: searchQueryType,
	})

	return utils.TemplRender(c, feed)
}
