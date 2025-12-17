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

	err := Client.Get(&movie, `
SELECT
    m.id,
    m.title,
    m.release_date,
    m.runtime,
    m.imdb_id,
    m.overview,
    m.original_title,
    m.tagline,
    se.name AS "series",
    se.id AS "series_id",
    ms.number_in_series,
    r.rating,
    COALESCE(ARRAY_TO_JSON(ARRAY (
                SELECT
                    jsonb_build_object('id', id, 'name', name)
                FROM ( SELECT DISTINCT ON (pc.id)
                    pc.id, pc.name FROM production_company pc
                    JOIN movie_company mc2 ON mc2.company_id = pc.id
                    WHERE
                        mc2.movie_id = m.id ORDER BY pc.id, pc.name) AS uniq_pc ORDER BY name ASC)), '[]') AS production_companies,
    COALESCE(ARRAY_TO_JSON(ARRAY (
                SELECT
                    jsonb_build_object('id', id, 'name', name)
                FROM ( SELECT DISTINCT ON (pc.id)
                        pc.id, pc.name
                    FROM production_country pc
                    JOIN movie_country mc2 ON mc2.country_id = pc.id
                    WHERE
                        mc2.movie_id = m.id ORDER BY pc.id, pc.name) AS uniq_pc ORDER BY name ASC)), '[]') AS production_countries,
    r.created_at at time zone 'UTC' at time zone 'Europe/Stockholm' AS "rated_at",
    COALESCE(ARRAY_TO_JSON(ARRAY_AGG(DISTINCT jsonb_build_object('name', g.name, 'id', g.id)) FILTER (WHERE g.name IS NOT NULL)), '[]') AS genres,
    COALESCE(ARRAY_TO_JSON(ARRAY_AGG(DISTINCT jsonb_build_object('name', l.english_name, 'id', l.id)) FILTER (WHERE l.english_name IS NOT NULL)), '[]') AS languages
FROM
    movie AS m
    LEFT JOIN movie_genre AS mg ON mg.movie_id = m.id
    LEFT JOIN genre AS g ON g.id = mg.genre_id
    LEFT JOIN rating AS r ON r.movie_id = m.id
        AND r.user_id = $2
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
    LEFT JOIN movie_language AS ml ON ml.movie_id = m.id
    LEFT JOIN "language" AS l ON l.id = ml.language_id
WHERE
    m.id = $1
GROUP BY
    1,
    r.id,
    se.id,
    ms.number_in_series
		`, mq.Id, mq.UserId)

	if err != nil {
		return types.Movie{}, err
	}

	return movie, nil
}

func (mq *MovieQueries) ReviewByID() (types.Review, error) {
	var review types.Review

	err := Client.Get(&review, `
SELECT
    id,
    content,
    private
FROM
    review
WHERE
    id = $1
		`, mq.Id)

	if err != nil {
		return review, err
	}

	return review, nil
}

func (mq *MovieQueries) ReviewByMovieID() (types.Review, error) {
	var review types.Review

	err := Client.Get(&review, `
SELECT
    id,
    content,
    private
FROM
    review
WHERE
    movie_id = $1
    AND user_id = $2
		`, mq.Id, mq.UserId)

	if err != nil {
		return review, err
	}

	return review, nil
}

func (mq *MovieQueries) RatingsByOthers() (types.OthersStats, error) {
	var others types.OthersStats

	err := Client.Get(&others, `
SELECT
    (
        SELECT
            count(DISTINCT user_id)
        FROM
            seen
        WHERE
            movie_id = $1) AS seen_count,
    (
        SELECT
            COALESCE(AVG(r.latest_rating), 0)
        FROM ( SELECT DISTINCT ON (user_id)
                user_id,
                rating AS latest_rating
            FROM
                rating
            WHERE
                movie_id = $1
            ORDER BY
                user_id,
                created_at DESC) r) AS avg_rating
`, mq.Id)

	if err != nil {
		return types.OthersStats{}, err
	}

	return others, nil
}

func (mq *MovieQueries) SeenByUser() ([]movie.WatchedAt, error) {
	var watchedAt []movie.WatchedAt

	err := Client.Select(&watchedAt, `
SELECT
    s.id,
    date at time zone 'UTC' at time zone 'Europe/Stockholm' AS date,
    COALESCE(ARRAY_AGG(u.name) FILTER (WHERE u.name IS NOT NULL), '{}') AS seen_with
FROM
    seen s
    LEFT JOIN seen_with sw ON sw.seen_id = s.id
    LEFT JOIN "user" u ON u.id = sw.other_user_id
WHERE
    movie_id = $1
    AND user_id = $2
GROUP BY
    1
ORDER BY
    date DESC
`, mq.Id, mq.UserId)

	if err != nil {
		return watchedAt, err
	}

	return watchedAt, nil
}

func (mq *MovieQueries) IsWatchlisted() (bool, error) {
	var isInWatchlist bool

	err := Client.Get(
		&isInWatchlist,
		`
		SELECT
		    EXISTS (
		        SELECT
		            *
		        FROM
		            watchlist
		        WHERE
		            movie_id = $1
		            AND user_id = $2)
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

	err := Client.Select(&castOrCrew, `
SELECT
    CASE mp.job
    WHEN 'cinematographer' THEN
        'Director of Photography'
    ELSE
        INITCAP(mp.job::text)
    END AS job,
    ARRAY_AGG(p.name ORDER BY num_movies DESC, p.popularity DESC, p.name ASC) AS people_names,
    ARRAY_AGG(p.id ORDER BY num_movies DESC, p.popularity DESC, p.name ASC) AS people_ids,
    CASE mp.job
    WHEN 'cast' THEN
        ARRAY_AGG(COALESCE(mp.character, '')
        ORDER BY num_movies DESC, p.popularity DESC, p.name ASC)
    ELSE
        ARRAY[]::text[]
    END AS characters
FROM
    movie_person AS mp
    INNER JOIN person AS p ON p.id = mp.person_id
    INNER JOIN (
        SELECT
            person_id,
            COUNT(*) AS num_movies
        FROM
            movie_person
        GROUP BY
            person_id) AS movie_counts ON p.id = movie_counts.person_id
WHERE
    mp.movie_id = $1
GROUP BY
    mp.job
    -- Sorts the cast and crew in a consistent order since UI renders
    -- it by looping through the array.
ORDER BY
    CASE mp.job
    WHEN 'director' THEN
        1
    WHEN 'writer' THEN
        2
    WHEN 'cast' THEN
        3
    WHEN 'composer' THEN
        4
    WHEN 'producer' THEN
        5
    WHEN 'cinematographer' THEN
        6
    WHEN 'editor' THEN
        7
    END
		`, mq.Id)

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
