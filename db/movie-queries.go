package db

import (
	"believer/movies/components/movie"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type MovieQueries struct {
	Id              string
	UserId          string
	IsAuthenticated bool
}

type CastDB struct {
	Job        string         `db:"job"`
	Names      pq.StringArray `db:"people_names"`
	Ids        pq.Int32Array  `db:"people_ids"`
	Characters pq.StringArray `db:"characters"`
}

func MakeMovieQueries(c *fiber.Ctx) (*MovieQueries, error) {
	movieId := c.Params("id")
	isAuthenticated := utils.IsAuthenticated(c)
	userId := c.Locals("UserId").(string)

	if movieId == "" {
		movieId = c.Query("id")
	}

	id, err := utils.SelfHealingUrl(movieId)

	if err != nil {
		return &MovieQueries{
			Id:              "0",
			IsAuthenticated: isAuthenticated,
			UserId:          userId,
		}, err
	}

	return &MovieQueries{
		Id:              id,
		IsAuthenticated: isAuthenticated,
		UserId:          userId,
	}, nil
}

func (mq *MovieQueries) GetByID() (types.Movie, error) {
	var movie types.Movie

	err := Dot.Get(Client, &movie, "movie-by-id", mq.Id, mq.UserId)

	if err != nil {
		return types.Movie{}, err
	}

	return movie, nil
}

func (mq *MovieQueries) ReviewByID() (types.Review, error) {
	var review types.Review

	err := Dot.Get(Client, &review, "review-by-movie-id", mq.Id, mq.UserId)

	if err != nil {
		return types.Review{}, err
	}

	return review, nil
}

func (mq *MovieQueries) RatingsByOthers() (types.OthersStats, error) {
	var others types.OthersStats

	err := Dot.Get(Client, &others, "others-ratings", mq.Id)

	if err != nil {
		return types.OthersStats{}, err
	}

	return others, nil
}

func (mq *MovieQueries) SeenByUser() ([]movie.WatchedAt, error) {
	var watchedAt []movie.WatchedAt

	if !mq.IsAuthenticated {
		return watchedAt, fiber.ErrUnauthorized
	}

	err := Dot.Select(Client, &watchedAt, "seen-by-user-id", mq.Id, mq.UserId)

	if err != nil {
		return watchedAt, err
	}

	return watchedAt, nil
}

func (mq *MovieQueries) IsWatchlisted() (bool, error) {
	var isInWatchlist bool

	err := Client.Get(
		&isInWatchlist,
		`SELECT
    EXISTS (
        SELECT
            *
        FROM
            watchlist
        WHERE
            movie_id = $1
            AND user_id = $2);
`,
		mq.Id,
		mq.UserId,
	)

	if err != nil {
		return false, err
	}

	return isInWatchlist, err
}

func (mq *MovieQueries) Cast() ([]views.CastDTO, bool, error) {
	var castOrCrew []CastDB

	err := Dot.Select(Client, &castOrCrew, "cast-by-id", mq.Id)

	if err != nil {
		return nil, false, err
	}

	updatedCastOrCrew := make([]views.CastDTO, len(castOrCrew))
	hasCharacters := false

	for i, cast := range castOrCrew {
		characters := cast.Characters

		if cast.Job == "Cast" {
			for _, value := range characters {
				if value != "" {
					hasCharacters = true
					break
				}
			}
		}

		if len(characters) == 0 {
			characters = make([]string, len(cast.Names))
		}

		updatedCastOrCrew[i] = views.CastDTO{
			Job:    cast.Job,
			People: zipCast(cast.Names, cast.Ids, characters),
		}
	}

	return updatedCastOrCrew, hasCharacters, nil
}

func zipCast(names []string, ids []int32, characters []string) []views.CastAndCrewDTO {
	zipped := make([]views.CastAndCrewDTO, len(names))

	for i := range names {
		zipped[i] = views.CastAndCrewDTO{
			Name:      names[i],
			ID:        ids[i],
			Character: characters[i],
		}
	}

	return zipped
}
