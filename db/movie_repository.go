package db

import (
	"believer/movies/components/list"
	"believer/movies/components/movie"
	"believer/movies/types"
	"believer/movies/views"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type MovieQuerier interface {
	// Movie Detail Queries
	GetByID(id string, userID string) (types.Movie, error)
	GetReviewByMovieID(movieID string, userID string) (types.Review, error)
	RatingsByOthers(id string) (types.OthersStats, error)
	SeenByUser(id string, userID string) ([]movie.WatchedAt, error)
	IsWatchlisted(id string, userID string) (bool, error)
	Cast(id string) ([]views.CastDTO, bool, error)

	// Seen/Watched Queries
	GetFriends(userID string) ([]list.DataListItem, error)
	GetSeenMovie(seenID string) (views.WatchData, error)
	DeleteSeenMovie(tx *sqlx.Tx, seenID string) error
	InsertSeenMovie(tx *sqlx.Tx, userID string, movieID int, watchedAt time.Time) (int, error)
	InsertSeenWith(tx *sqlx.Tx, seenID int, friends []string) error
	CreateSeenMovieDirect(userID string, movieID string) error
	UpdateSeenMovie(seenID string, watchedAt time.Time, friends []string) error

	// Ratings Queries
	DeleteRating(tx *sqlx.Tx, movieID int, userID string) error
	AddRating(tx *sqlx.Tx, userID string, movieID int, rating int) error
	UpdateRating(userID string, movieID int, rating string) error

	// General Movie Queries
	GetMovieByIDSimple(movieID int) (types.Movie, error)
	GetMovieByImdbID(imdbID string) (types.Movie, error)
	GetMovieTitleAndImdbID(movieID string) (types.Movie, error)
	GetUserWatchProviders(userID string) (string, error)
	GetAllSeries() ([]list.DataListItem, error)
	GetMovieAwards(imdbID string, awardType string) ([]types.Award, error)

	// Create/Update Movie from post form
	InsertReview(tx *sqlx.Tx, content string, private bool, userID string, movieID int) error
	GetOrInsertSeries(tx *sqlx.Tx, name string) (int, error)
	InsertMovieSeries(tx *sqlx.Tx, movieID int, seriesID int, numberInSeries int) error
	InsertWatchlist(tx *sqlx.Tx, userID string, movieID int) error
	DeleteWatchlist(tx *sqlx.Tx, userID string, movieID int) error
	DeleteNowPlaying(tx *sqlx.Tx, userID string, imdbID string) error
	UpdateMovie(tx *sqlx.Tx, id int, title string, runtime int, releaseDate string, imdbID string, overview string, poster string, tagline string, tmdbID int) error
	UpdateNowPlaying(imdbID string, position float64, userID string) error
	MovieExists(imdbID string) (bool, error)
	Begin() (*sqlx.Tx, error)
}

type MovieRepository struct {
	db *sqlx.DB
}

func NewMovieRepository(db *sqlx.DB) *MovieRepository {
	return &MovieRepository{db}
}

func (r *MovieRepository) Begin() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

type CastDB struct {
	Job        string         `db:"job"`
	Names      pq.StringArray `db:"people_names"`
	Ids        pq.Int32Array  `db:"people_ids"`
	Characters pq.StringArray `db:"characters"`
}

// Movie Detail Queries
// =====================================================

func (r *MovieRepository) GetByID(id string, userID string) (types.Movie, error) {
	var movie types.Movie

	err := r.db.Get(&movie, `
SELECT
    m.id,
    m.title,
    m.release_date,
    m.runtime,
    m.imdb_id,
    m.overview,
    m.original_title,
    m.tagline,
    MIN(se.name) AS "series",
    MIN(se.id) AS "series_id",
    MIN(ms.number_in_series) AS "number_in_series",
    COALESCE(ARRAY_TO_JSON(ARRAY (
                SELECT
                    jsonb_build_object('id', s2.id::text, 'name', s2.name, 'number_in_series', ms2.number_in_series)
                FROM movie_series ms2
                JOIN series s2 ON s2.id = ms2.series_id
                WHERE
                    ms2.movie_id = m.id ORDER BY s2.name ASC)), '[]') AS all_series,
    r.rating,
    COALESCE(ARRAY_TO_JSON(ARRAY (
                SELECT
                    jsonb_build_object('id', id::text, 'name', name)
                FROM ( SELECT DISTINCT ON (pc.id)
                        pc.id, pc.name
                    FROM production_company pc
                    JOIN movie_company mc2 ON mc2.company_id = pc.id
                    WHERE
                        mc2.movie_id = m.id ORDER BY pc.id, pc.name) AS uniq_pc ORDER BY name ASC)), '[]') AS production_companies,
    COALESCE(ARRAY_TO_JSON(ARRAY (
                SELECT
                    jsonb_build_object('id', id::text, 'name', name)
            FROM ( SELECT DISTINCT ON (pc.id)
                    pc.id, pc.name
                FROM production_country pc
                JOIN movie_country mc2 ON mc2.country_id = pc.id
                WHERE
                    mc2.movie_id = m.id ORDER BY pc.id, pc.name) AS uniq_pc ORDER BY name ASC)), '[]') AS production_countries,
    r.created_at at time zone 'UTC' at time zone 'Europe/Stockholm' AS "rated_at",
    COALESCE(ARRAY_TO_JSON(ARRAY_AGG(DISTINCT jsonb_build_object('name', g.name, 'id', g.id::text)) FILTER (WHERE g.name IS NOT NULL)), '[]') AS genres,
    COALESCE(ARRAY_TO_JSON(ARRAY_AGG(DISTINCT jsonb_build_object('name', l.english_name, 'id', l.id::text)) FILTER (WHERE l.english_name IS NOT NULL)), '[]') AS languages
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
    r.id
		`, id, userID)

	return movie, err
}

func (r *MovieRepository) GetReviewByMovieID(movieID string, userID string) (types.Review, error) {
	var review types.Review

	err := r.db.Get(&review, `
SELECT
    id,
    content,
    private
FROM
    review
WHERE
    movie_id = $1
    AND user_id = $2
		`, movieID, userID)

	return review, err
}

func (r *MovieRepository) RatingsByOthers(id string) (types.OthersStats, error) {
	var others types.OthersStats

	err := r.db.Get(&others, `
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
`, id)

	return others, err
}

func (r *MovieRepository) SeenByUser(id string, userID string) ([]movie.WatchedAt, error) {
	var watchedAt []movie.WatchedAt

	err := r.db.Select(&watchedAt, `
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
`, id, userID)

	return watchedAt, err
}

func (r *MovieRepository) IsWatchlisted(id string, userID string) (bool, error) {
	var isInWatchlist bool

	err := r.db.Get(
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
		id,
		userID,
	)

	return isInWatchlist, err
}

func (r *MovieRepository) Cast(id string) ([]views.CastDTO, bool, error) {
	var castOrCrew []CastDB

	err := r.db.Select(&castOrCrew, `
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
    -- Same as in movie-queries.go
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
		`, id)

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

// Seen/Watched Queries
// =====================================================

func (r *MovieRepository) GetFriends(userID string) ([]list.DataListItem, error) {
	var friends []list.DataListItem
	err := r.db.Select(&friends, `
		SELECT
		    id AS "value",
		    name
		FROM
		    "user"
		WHERE
		    id != $1
		`, userID)
	return friends, err
}

func (r *MovieRepository) GetSeenMovie(seenID string) (views.WatchData, error) {
	var watch views.WatchData
	err := r.db.Get(&watch, `
SELECT
    TO_CHAR(date AT TIME ZONE 'UTC' AT TIME ZONE 'Europe/Stockholm', 'YYYY-MM-DD"T"HH24:MI') AS date,
    COALESCE(ARRAY_AGG(sw.other_user_id) FILTER (WHERE sw.other_user_id IS NOT NULL), '{}') AS seen_with
FROM
    seen s
    LEFT JOIN seen_with sw ON sw.seen_id = s.id
WHERE
    id = $1
GROUP BY
    id
`, seenID)
	return watch, err
}

func (r *MovieRepository) DeleteSeenMovie(tx *sqlx.Tx, seenID string) error {
	_, err := tx.Exec(`DELETE FROM seen_with
WHERE seen_id = $1`, seenID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM seen
WHERE id = $1`, seenID)
	return err
}

func (r *MovieRepository) InsertSeenMovie(tx *sqlx.Tx, userID string, movieID int, watchedAt time.Time) (int, error) {
	var id int
	err := tx.Get(&id, `
INSERT INTO seen (user_id, movie_id, date)
    VALUES ($1, $2, $3)
RETURNING
    id
	`, userID, movieID, watchedAt)
	return id, err
}

func (r *MovieRepository) InsertSeenWith(tx *sqlx.Tx, seenID int, friends []string) error {
	_, err := tx.Exec(`
		INSERT INTO seen_with (seen_id, other_user_id)
		SELECT
		    $1,
		    UNNEST($2::text[])::int
		ON CONFLICT
		    DO NOTHING
	`, seenID, pq.Array(friends))
	return err
}

func (r *MovieRepository) CreateSeenMovieDirect(userID string, movieID string) error {
	_, err := r.db.Exec(`INSERT INTO seen (user_id, movie_id)
    VALUES ($1, $2)`, userID, movieID)
	return err
}

func (r *MovieRepository) UpdateSeenMovie(seenID string, watchedAt time.Time, friends []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE
    seen
SET
    date = $1
WHERE
    id = $2`, watchedAt, seenID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM seen_with
WHERE seen_id = $1`, seenID)
	if err != nil {
		return err
	}

	if len(friends) > 0 {
		_, err = tx.Exec(`
			INSERT INTO seen_with (seen_id, other_user_id)
			SELECT
			    $1,
			    UNNEST($2::text[])::int
			ON CONFLICT
			    DO NOTHING
		`, seenID, pq.Array(friends))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Ratings Queries
// =====================================================

func (r *MovieRepository) DeleteRating(tx *sqlx.Tx, movieID int, userID string) error {
	_, err := tx.Exec(`DELETE FROM rating
WHERE movie_id = $1
    AND user_id = $2`, movieID, userID)
	return err
}

func (r *MovieRepository) AddRating(tx *sqlx.Tx, userID string, movieID int, rating int) error {
	_, err := tx.Exec(`INSERT INTO rating (user_id, movie_id, rating)
    VALUES ($1, $2, $3)`, userID, movieID, rating)
	return err
}

func (r *MovieRepository) UpdateRating(userID string, movieID int, rating string) error {
	_, err := r.db.Exec(`
UPDATE
    rating
SET
    rating = $1,
    updated_at = NOW()
WHERE
    movie_id = $2
    AND user_id = $3`, rating, movieID, userID)
	return err
}

// General Movie Queries
// =====================================================

func (r *MovieRepository) GetMovieByIDSimple(movieID int) (types.Movie, error) {
	var movie types.Movie
	err := r.db.Get(&movie, `SELECT
    id,
    title
FROM
    movie
WHERE
    id = $1`, movieID)
	return movie, err
}

func (r *MovieRepository) GetMovieByImdbID(imdbID string) (types.Movie, error) {
	var movie types.Movie
	err := r.db.Get(&movie, "SELECT title, id FROM movie WHERE imdb_id = $1", imdbID)
	return movie, err
}

func (r *MovieRepository) GetMovieTitleAndImdbID(movieID string) (types.Movie, error) {
	var movie types.Movie
	err := r.db.Get(&movie, `
SELECT
    imdb_id,
    title
FROM
    movie
WHERE
    id = $1`, movieID)
	return movie, err
}

func (r *MovieRepository) GetUserWatchProviders(userID string) (string, error) {
	var storedProviders string
	err := r.db.Get(&storedProviders, `
SELECT
    watch_providers
FROM
    "user"
WHERE
    id = $1`, userID)
	return storedProviders, err
}

func (r *MovieRepository) GetAllSeries() ([]list.DataListItem, error) {
	var options []list.DataListItem
	err := r.db.Select(&options, `SELECT
    id AS "value",
    name
FROM
    series
ORDER BY
    name ASC`)
	return options, err
}

func (r *MovieRepository) GetMovieAwards(imdbID string, awardType string) ([]types.Award, error) {
	var awards []types.Award
	err := r.db.Select(&awards, `
SELECT
    name AS category,
    year,
    COALESCE(JSONB_AGG(
            CASE WHEN person IS NOT NULL
                AND person_id IS NOT NULL THEN
                JSONB_BUILD_OBJECT('name', person, 'id', person_id)
            WHEN person IS NOT NULL THEN
                JSONB_BUILD_OBJECT('name', person)
            ELSE
                JSONB_BUILD_OBJECT('name', 'N/A')
            END) FILTER (WHERE person IS NOT NULL
            OR person_id IS NOT NULL), '[]'::jsonb) AS nominees,
    winner,
    detail
FROM
    award
WHERE
    imdb_id = $1
    AND type = $2
GROUP BY
    name,
    year,
    winner,
    detail
ORDER BY
    winner DESC,
    category ASC
		`, imdbID, awardType)
	return awards, err
}

// Create/Update Movie queries
// =====================================================

func (r *MovieRepository) InsertReview(tx *sqlx.Tx, content string, private bool, userID string, movieID int) error {
	_, err := tx.Exec(`
INSERT INTO review (content, private, user_id, movie_id)
    VALUES ($1, $2, $3, $4)
ON CONFLICT
    DO NOTHING
	`, content, private, userID, movieID)
	return err
}

func (r *MovieRepository) GetOrInsertSeries(tx *sqlx.Tx, name string) (int, error) {
	var seriesID int
	err := tx.Get(&seriesID, `
INSERT INTO series (name)
    VALUES ($1)
ON CONFLICT
    DO UPDATE SET
        name = EXCLUDED.name
    RETURNING
        id`, name)
	return seriesID, err
}

func (r *MovieRepository) InsertMovieSeries(tx *sqlx.Tx, movieID int, seriesID int, numberInSeries int) error {
	_, err := tx.Exec(`
INSERT INTO movie_series (movie_id, series_id, number_in_series)
    VALUES ($1, $2, $3)`, movieID, seriesID, numberInSeries)
	return err
}

func (r *MovieRepository) InsertWatchlist(tx *sqlx.Tx, userID string, movieID int) error {
	_, err := tx.Exec(`
INSERT INTO watchlist (user_id, movie_id)
    VALUES ($1, $2)`, userID, movieID)
	return err
}

func (r *MovieRepository) DeleteWatchlist(tx *sqlx.Tx, userID string, movieID int) error {
	_, err := tx.Exec(`
DELETE FROM watchlist
WHERE user_id = $1
    AND movie_id = $2`, userID, movieID)
	return err
}

func (r *MovieRepository) DeleteNowPlaying(tx *sqlx.Tx, userID string, imdbID string) error {
	_, err := tx.Exec(`
DELETE FROM now_playing
WHERE user_id = $1
    AND imdb_id = $2`, userID, imdbID)
	return err
}

func (r *MovieRepository) UpdateMovie(tx *sqlx.Tx, id int, title string, runtime int, releaseDate string, imdbID string, overview string, poster string, tagline string, tmdbID int) error {
	_, err := tx.Exec(`
UPDATE
    movie
SET
    title = $2,
    runtime = $3,
    release_date = NULLIF ($4, '')::date,
    imdb_id = $5,
    overview = $6,
    poster = $7,
    tagline = $8,
    updated_at = NOW(),
    tmdb_id = $9
WHERE
    id = $1;

`, id, title, runtime, releaseDate, imdbID, overview, poster, tagline, tmdbID)
	return err
}

func (r *MovieRepository) UpdateNowPlaying(imdbID string, position float64, userID string) error {
	_, err := r.db.Exec(`
		INSERT INTO now_playing (imdb_id, position, user_id)
		    VALUES ($1, $2, $3)
		ON CONFLICT (imdb_id, user_id)
		    DO UPDATE SET
		        position = excluded.position
	`, imdbID, position, userID)
	return err
}

func (r *MovieRepository) MovieExists(imdbID string) (bool, error) {
	var movieExists bool
	err := r.db.Get(&movieExists, `
		SELECT
		    EXISTS (
		        SELECT
		            1
		        FROM
		            movie
		        WHERE
		            imdb_id = $1)
	`, imdbID)
	return movieExists, err
}
