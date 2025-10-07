package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetFeed(c *fiber.Ctx) error {
	var movies types.Movies
	var persons types.Persons

	page := c.QueryInt("page", 1)
	lastHeader := c.Query("last-header", "0000-00-January")
	searchQuery := c.Query("search")
	userId := c.Locals("UserId")
	searchQueryType := "movie"

	if searchQuery != "" {
		c.Set("HX-Push-Url", fmt.Sprintf("/?search=%s", searchQuery))

		// Query string with a specifier for type. For example:
		// - movie:godfa
		// - actor:ryan
		// - rating:3
		if queryType, query, ok := strings.Cut(searchQuery, ":"); ok {
			queryType = strings.ToLower(queryType)
			query = strings.TrimSpace(query)

			if queryType == "dp" || queryType == "dop" {
				queryType = "cinematographer"
			}

			switch queryType {
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
			case "director", "writer", "producer", "composer", "cinematographer", "editor":
				err := db.Dot.Select(db.Client, &persons, "feed-search-job", query, queryType)
				searchQueryType = "person"

				if err != nil {
					return err
				}
			case "rating":
				err := db.Dot.Select(db.Client, &movies, "feed-search-rating", query, userId)

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
		err := db.Dot.Select(db.Client, &movies, "feed", (page-1)*20, userId)

		c.Set("HX-Push-Url", "/")

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

	// Group movies for display by year and month
	groupedMovies := make(map[string]types.Movies)

	for _, m := range movies {
		key := m.WatchedAt.Format("2006-01-January")
		groupedMovies[key] = append(groupedMovies[key], m)
	}

	// Grouping is not sorted. Get and sort the keys in descending order
	// and use the keys for presentation
	keys := make([]string, 0, len(groupedMovies))
	for k := range groupedMovies {
		keys = append(keys, k)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	feed := views.Feed(views.FeedProps{
		IsAdmin:       utils.IsAuthenticated(c),
		LastHeader:    lastHeader,
		GroupedMovies: groupedMovies,
		SortedKeys:    keys,
		NextPage:      page + 1,
		Persons:       persons,
		Query:         searchQuery,
		QueryType:     searchQueryType,
	})

	return utils.Render(c, feed)
}
