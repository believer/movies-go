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

	querySearch := `
SELECT
    m.id,
    m.title,
    m.overview,
    se.name AS "series",
    ms.number_in_series,
    m.release_date AS watched_at
FROM
    movie AS m
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
WHERE
    m.title ILIKE '%' || $1 || '%'
    OR m.original_title ILIKE '%' || $1 || '%'
    OR se.name ILIKE '%' || $1 || '%'
ORDER BY
    m.release_date DESC
			`

	queryJob := `
SELECT
    p.id,
    p.name,
    count(*)
FROM
    person p
    INNER JOIN movie_person mp ON mp.person_id = p.id
WHERE
    p."name" ILIKE '%' || $1 || '%'
    AND mp.job = $2
GROUP BY
    p.id
ORDER BY
    COUNT DESC
LIMIT 100
			`

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
				err := db.Client.Select(&movies, querySearch, query)

				if err != nil {
					return err
				}
			case "actor", "cast":
				err := db.Client.Select(&persons, queryJob, query, "cast")
				searchQueryType = "person"

				if err != nil {
					return err
				}
			case "director", "writer", "producer", "composer", "cinematographer", "editor":
				err := db.Client.Select(&persons, queryJob, query, queryType)
				searchQueryType = "person"

				if err != nil {
					return err
				}
			case "rating":
				err := db.Client.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.overview,
    se.name AS "series",
    ms.number_in_series,
    m.release_date AS watched_at
FROM
    movie AS m
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
    LEFT JOIN rating AS r ON r.movie_id = m.id
WHERE
    r.rating = $1
    AND r.user_id = $2
ORDER BY
    m.release_date DESC
					`, query, userId)

				if err != nil {
					return err
				}
			}
		} else {
			err := db.Client.Select(&movies, querySearch, searchQuery)

			if err != nil {
				return err
			}
		}
	} else {
		err := db.Client.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.overview,
    m.release_date,
    se.name AS "series",
    ms.number_in_series,
    s.date at time zone 'UTC' at time zone 'Europe/Stockholm' AS watched_at
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
WHERE
    user_id = $2
ORDER BY
    s.date DESC OFFSET $1
LIMIT 20
			`, (page-1)*20, userId)

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
		key := "TBD"

		if m.WatchedAt.Valid {
			key = m.WatchedAt.Time.Format("2006-01-January")
		}

		if searchQuery != "" && m.ReleaseDate.Valid {
			key = m.ReleaseDate.Time.Format("2006-01-January")
		}

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
