package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type FeedHandler struct {
	repo           db.FeedQuerier
	nowPlayingRepo db.NowPlayingQuerier
}

func NewFeedHandler(repo db.FeedQuerier, nowPlayingRepo db.NowPlayingQuerier) *FeedHandler {
	return &FeedHandler{repo, nowPlayingRepo}
}

func (h *FeedHandler) GetFeed(c *fiber.Ctx) error {
	var movies types.Movies
	var persons types.Persons
	var nowPlaying types.Movies

	page := c.QueryInt("page", 1)
	lastHeader := c.Query("last-header", "0000-00-January")
	searchQuery := c.Query("search")
	userID := c.Locals("UserId").(string)
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
				var err error
				movies, err = h.repo.SearchMovies(query)

				if err != nil {
					return err
				}
			case "actor", "cast":
				var err error
				persons, err = h.repo.SearchPersons(query, "cast")
				searchQueryType = "person"

				if err != nil {
					return err
				}
			case "director", "writer", "producer", "composer", "cinematographer", "editor":
				var err error
				persons, err = h.repo.SearchPersons(query, queryType)
				searchQueryType = "person"

				if err != nil {
					return err
				}
			case "rating":
				var err error
				movies, err = h.repo.SearchMoviesByRating(query, userID)

				if err != nil {
					return err
				}
			}
		} else {
			var err error
			movies, err = h.repo.SearchMovies(searchQuery)

			if err != nil {
				return err
			}
		}
	} else {
		var err error
		movies, err = h.repo.GetFeedMovies(userID, (page-1)*20)

		if err != nil {
			return err
		}

		c.Set("HX-Push-Url", "/")
	}

	var err error
	nowPlaying, err = h.nowPlayingRepo.GetNowPlaying(userID)

	if err != nil {
		slog.Error("[Now Playing]", "error", err)
		return err
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
		NowPlaying:    nowPlaying,
		Persons:       persons,
		Query:         searchQuery,
		QueryType:     searchQueryType,
	})

	return utils.Render(c, feed)
}
