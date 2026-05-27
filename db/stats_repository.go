package db

import (
	"believer/movies/components/graph"
	"believer/movies/types"

	"github.com/jmoiron/sqlx"
)

type StatsQuerier interface {
	GetReviewsCount(userID string) (int, error)
	GetStatsData(userID string) (types.Stats, error)
	GetMostWatchedByJob(job string, userID string, year string) ([]types.ListItem, error)
	GetMostWatchedMovies(userID string) ([]types.ListItem, error)
	GetMoviesByYear(userID string, year string) ([]graph.GraphData, error)
	GetRatings(userID string) ([]graph.GraphData, error)
	GetWatchedThisYearByMonth(userID string, yearTime string) ([]graph.GraphData, error)
	GetShortestAndLongestMovie(userID string) (types.Movies, error)
	GetTotalWatchedByJobAndYear(userID string, job string, year string) ([]types.ListItem, error)
	GetWatchedByYear(userID string) ([]graph.GraphData, error)
	GetWatchedByWeekday(userID string, yearTime string) ([]graph.GraphData, error)
	GetWilhelmScreamCount(userID string) ([]int, error)
	GetRatingsThisYear(userID string, yearTime string) ([]graph.GraphData, error)
	GetMostAwardNominations(userID string) (types.AwardPersonStat, error)
	GetMostAwardWins(userID string) (types.AwardPersonStat, error)
	GetTopAwardedMovies(userID string) ([]types.AwardMovieStat, error)
	GetRatingsForYear(year string, userID string) (types.Movies, error)
	GetBestOfTheYear(userID string, year string) ([]types.ListItem, error)
	GetWilhelmMovies(userID string, offset int) (types.Movies, error)
	GetSeenWith() ([]types.ListItem, error)
	GetHighestRankedPersonByJob(userID string, job string) ([]types.HighestRated, error)
}

type StatsRepository struct {
	db *sqlx.DB
}

func NewStatsRepository(db *sqlx.DB) *StatsRepository {
	return &StatsRepository{db}
}

func (r *StatsRepository) GetReviewsCount(userID string) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT
    count(*)
FROM
    review
WHERE
    user_id = $1`, userID)
	return count, err
}

func (r *StatsRepository) GetStatsData(userID string) (types.Stats, error) {
	var stats types.Stats
	err := r.db.Get(&stats, `
SELECT
    COUNT(DISTINCT movie_id) AS unique_movies,
    COUNT(movie_id) seen_with_rewatches,
    COALESCE(SUM(m.runtime), 0) AS total_runtime
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
WHERE
    user_id = $1
	`, userID)
	return stats, err
}

func (r *StatsRepository) GetMostWatchedByJob(job string, userID string, year string) ([]types.ListItem, error) {
	var items []types.ListItem
	err := r.db.Select(&items, `
WITH aggregated_cast AS (
    SELECT
        mp.person_id,
        COUNT(*) AS count
    FROM ( SELECT DISTINCT ON (movie_id)
            movie_id
        FROM
            seen
        WHERE
            user_id = $2
            AND ($3 = 'All'
                OR EXTRACT(YEAR FROM date) = $3::int)) AS s
        INNER JOIN movie_person AS mp ON mp.movie_id = s.movie_id
    WHERE
        mp.job = $1
    GROUP BY
        mp.person_id
    ORDER BY
        count DESC
    LIMIT 50
)
SELECT
    ac.count,
    p.name,
    p.id
FROM
    aggregated_cast ac
    INNER JOIN person p ON p.id = ac.person_id
ORDER BY
    ac.count DESC,
    p.name ASC
LIMIT 10
	`, job, userID, year)
	return items, err
}

func (r *StatsRepository) GetMostWatchedMovies(userID string) ([]types.ListItem, error) {
	var items []types.ListItem
	err := r.db.Select(&items, `
SELECT
    COUNT(*) AS count,
    m.title AS name,
    m.id
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
WHERE
    user_id = $1
GROUP BY
    m.id
ORDER BY
    count DESC
LIMIT 20
	`, userID)
	return items, err
}

func (r *StatsRepository) GetMoviesByYear(userID string, year string) ([]graph.GraphData, error) {
	var data []graph.GraphData
	err := r.db.Select(&data, `
SELECT
    EXTRACT(YEAR FROM release_date) AS label,
    COUNT(*) AS value
FROM ( SELECT DISTINCT
        movie_id
    FROM
        seen
    WHERE
        user_id = $1
        AND ($2 = 'All'
            OR EXTRACT(YEAR FROM date) = $2::int)) AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
GROUP BY
    label
ORDER BY
    label DESC
	`, userID, year)
	return data, err
}

func (r *StatsRepository) GetRatings(userID string) ([]graph.GraphData, error) {
	var data []graph.GraphData
	err := r.db.Select(&data, `
SELECT
    COUNT(*) AS value,
    rating AS label
FROM
    rating
WHERE
    user_id = $1
GROUP BY
    rating
ORDER BY
    rating
	`, userID)
	return data, err
}

func (r *StatsRepository) GetWatchedThisYearByMonth(userID string, yearTime string) ([]graph.GraphData, error) {
	var data []graph.GraphData
	err := r.db.Select(&data, `
WITH months (
    month
) AS (
    SELECT
        generate_series(DATE_TRUNC('year', $2::date), DATE_TRUNC('year', $2::date) + INTERVAL '1 year' - INTERVAL '1 day', INTERVAL '1 month'))
SELECT
    TO_CHAR(months.month, 'Mon') AS label,
    COALESCE(count(seen.id), 0) AS value
FROM
    months
    LEFT JOIN seen ON DATE_TRUNC('month', seen.date) = months.month
        AND user_id = $1
WHERE
    EXTRACT(YEAR FROM seen.date) = EXTRACT(YEAR FROM $2::date)
    OR seen.date IS NULL
GROUP BY
    months.month
ORDER BY
    months.month
	`, userID, yearTime)
	return data, err
}

func (r *StatsRepository) GetShortestAndLongestMovie(userID string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, `
(
    SELECT
        m.id,
        m.title,
        m.runtime
    FROM
        movie m
        JOIN seen s ON m.id = s.movie_id
    WHERE
        s.user_id = $1
    ORDER BY
        m.runtime ASC
    LIMIT 1)
UNION ALL (
    SELECT
        m.id,
        m.title,
        m.runtime
    FROM
        movie m
        JOIN seen s ON m.id = s.movie_id
    WHERE
        s.user_id = $1
    ORDER BY
        m.runtime DESC
    LIMIT 1)
	`, userID)
	return movies, err
}

func (r *StatsRepository) GetTotalWatchedByJobAndYear(userID string, job string, year string) ([]types.ListItem, error) {
	var items []types.ListItem
	err := r.db.Select(&items, `
SELECT
    COUNT(*) AS count
FROM
    seen s
    INNER JOIN movie_person mp ON mp.movie_id = s.movie_id
WHERE
    user_id = $1
    AND mp.job = $2
    AND ($3 = 'All'
        OR EXTRACT(YEAR FROM date) = $3::int)
	`, userID, job, year)
	return items, err
}

func (r *StatsRepository) GetWatchedByYear(userID string) ([]graph.GraphData, error) {
	var data []graph.GraphData
	err := r.db.Select(&data, `
SELECT
    EXTRACT(YEAR FROM date) AS label,
    COUNT(*) AS value
FROM
    seen
WHERE
    user_id = $1
    -- 2011 is where all the data that I hadn't tracked
    -- before I started ended up. So, there's a bunch of
    -- movies that year.
    AND EXTRACT(YEAR FROM date) > 2011
GROUP BY
    label
ORDER BY
    label
	`, userID)
	return data, err
}

func (r *StatsRepository) GetWatchedByWeekday(userID string, yearTime string) ([]graph.GraphData, error) {
	var data []graph.GraphData
	err := r.db.Select(&data, `
WITH days (
    day_val
) AS (
    SELECT
        generate_series(1, 7))
SELECT
    TRIM(TO_CHAR(make_date(2023, 1, 1) + (days.day_val * INTERVAL '1 day'), 'Day')) AS label,
    COALESCE(count(seen.id), 0) AS value
FROM
    days
    LEFT JOIN seen ON EXTRACT(ISODOW FROM seen.date) = days.day_val
        AND user_id = $1
        AND ($2 = 'All'
            OR EXTRACT(YEAR FROM seen.date) = EXTRACT(YEAR FROM $2::date))
        AND EXTRACT(YEAR FROM seen.date) > 2011
GROUP BY
    days.day_val
ORDER BY
    days.day_val
	`, userID, yearTime)
	return data, err
}

func (r *StatsRepository) GetWilhelmScreamCount(userID string) ([]int, error) {
	var wilhelms []int
	err := r.db.Select(&wilhelms, `
WITH seen_once AS (
    SELECT DISTINCT
        movie_id
    FROM
        seen s
    WHERE
        user_id = $1
)
SELECT
    COUNT(*)
FROM
    seen_once s
    INNER JOIN movie m ON m.id = s.movie_id
        AND m.wilhelm = TRUE
	`, userID)
	return wilhelms, err
}

func (r *StatsRepository) GetRatingsThisYear(userID string, yearTime string) ([]graph.GraphData, error) {
	var data []graph.GraphData
	err := r.db.Select(&data, `
WITH rating_series AS (
    SELECT
        generate_series(1, 10) AS rating_value
)
SELECT
    rs.rating_value AS label,
    COUNT(
        CASE WHEN s.movie_id IS NOT NULL THEN
            r.movie_id
        ELSE
            NULL
        END) AS value
FROM
    rating_series rs
    LEFT JOIN rating r ON r.rating = rs.rating_value
        AND r.user_id = $1
    LEFT JOIN seen s ON s.movie_id = r.movie_id
        AND s.user_id = $1
        AND EXTRACT(YEAR FROM s.date) = EXTRACT(YEAR FROM $2::date)
GROUP BY
    rs.rating_value
ORDER BY
    rs.rating_value
	`, userID, yearTime)
	return data, err
}

func (r *StatsRepository) GetMostAwardNominations(userID string) (types.AwardPersonStat, error) {
	var stat types.AwardPersonStat
	err := r.db.Get(&stat, `
WITH seen_movies AS (
    SELECT
        movie_id
    FROM
        seen
    WHERE
        user_id = $1
)
SELECT
    count(*) AS COUNT,
    a.person,
    a.person_id
FROM
    award a
WHERE
    EXISTS (
        SELECT
            1
        FROM
            movie_person mp
        WHERE
            mp.person_id = a.person_id
            AND mp.movie_id IN (
                SELECT
                    movie_id
                FROM
                    seen_movies))
    GROUP BY
        a.person_id,
        person
    ORDER BY
        COUNT DESC
    LIMIT 1
	`, userID)
	return stat, err
}

func (r *StatsRepository) GetMostAwardWins(userID string) (types.AwardPersonStat, error) {
	var stat types.AwardPersonStat
	err := r.db.Get(&stat, `
WITH seen_movies AS (
    SELECT
        movie_id
    FROM
        seen
    WHERE
        user_id = $1
)
SELECT
    count(*) AS COUNT,
    a.person,
    a.person_id
FROM
    award a
WHERE
    winner = TRUE
    AND EXISTS (
        SELECT
            1
        FROM
            movie_person mp
        WHERE
            mp.person_id = a.person_id
            AND mp.movie_id IN (
                SELECT
                    movie_id
                FROM
                    seen_movies))
    GROUP BY
        a.person_id,
        person
    ORDER BY
        COUNT DESC
    LIMIT 1
	`, userID)
	return stat, err
}

func (r *StatsRepository) GetTopAwardedMovies(userID string) ([]types.AwardMovieStat, error) {
	var movies []types.AwardMovieStat
	err := r.db.Select(&movies, `
WITH movie_awards AS (
    SELECT
        m.id,
        m.title,
        COUNT(DISTINCT (a.name, a.type)) AS award_count
    FROM
        seen s
        INNER JOIN movie m ON m.id = s.movie_id
        INNER JOIN award a ON a.imdb_id = m.imdb_id
    WHERE
        s.user_id = $1
        AND a.winner = TRUE
    GROUP BY
        m.id,
        m.title
)
SELECT
    *
FROM
    movie_awards
WHERE
    award_count = (
        SELECT
            MAX(award_count)
        FROM
            movie_awards)
	`, userID)
	return movies, err
}

func (r *StatsRepository) GetRatingsForYear(year string, userID string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, `
WITH RankedRatings AS (
    SELECT
        r.movie_id,
        r.user_id,
        r.rating,
        r.created_at,
        ROW_NUMBER() OVER (PARTITION BY r.movie_id,
            r.user_id ORDER BY r.created_at DESC) AS rn
    FROM
        rating r
)
SELECT DISTINCT
    m.id,
    m.title,
    rr.rating,
    s.date AS "watched_at"
FROM
    seen s
    INNER JOIN movie m ON m.id = s.movie_id
    INNER JOIN RankedRatings rr ON rr.movie_id = s.movie_id
        AND rr.user_id = s.user_id
WHERE
    EXTRACT('YEAR' FROM s.date) = $1
    AND s.user_id = $2
    AND rr.rn = 1
ORDER BY
    rr.rating DESC,
    s.date ASC
	`, year, userID)
	return movies, err
}

func (r *StatsRepository) GetBestOfTheYear(userID string, year string) ([]types.ListItem, error) {
	var movies []types.ListItem
	err := r.db.Select(&movies, `
WITH max_rating AS (
    SELECT DISTINCT
        s.movie_id,
        s.user_id,
        r.rating
    FROM
        seen s
        INNER JOIN rating r ON r.movie_id = s.movie_id
            AND r.user_id = $1
    WHERE
        s.user_id = $1
        AND date >= make_date($2, 1, 1)
        AND date < make_date($2 + 1, 1, 1) -- Seen in the given year
    GROUP BY
        s.id,
        r.rating
    HAVING
        COUNT(*) = 1 -- Seen exactly once in the given year
        AND s.movie_id NOT IN (
            SELECT
                movie_id
            FROM
                seen
            WHERE
                user_id = $1
                AND date < make_date($2, 1, 1) -- Seen before the given year
                OR date >= make_date($2 + 1, 1, 1) -- Seen after the given year
))
SELECT
    m.title AS "name",
    m.id AS "id",
    mr.rating AS "count"
FROM
    max_rating mr
    INNER JOIN movie m ON m.id = mr.movie_id
WHERE
    rating = (
        SELECT
            max(rating)
        FROM
            max_rating)
	`, userID, year)
	return movies, err
}

func (r *StatsRepository) GetWilhelmMovies(userID string, offset int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, `
WITH seen_once AS (
    SELECT DISTINCT
        movie_id
    FROM
        seen s
    WHERE
        user_id = $1
)
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    seen_once so
    INNER JOIN movie m ON m.id = so.movie_id
    LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            public.seen
        WHERE
            user_id = $1
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id
WHERE
    m.wilhelm = TRUE
ORDER BY
    m.release_date DESC OFFSET $2
LIMIT 50
	`, userID, offset)
	return movies, err
}

func (r *StatsRepository) GetSeenWith() ([]types.ListItem, error) {
	var items []types.ListItem
	err := r.db.Select(&items, `
SELECT
    u.name,
    COUNT(*)
FROM
    seen s
    RIGHT JOIN seen_with sw ON sw.seen_id = s.id
    INNER JOIN "user" u ON u.id = sw.other_user_id
WHERE
    user_id = 1
GROUP BY
    1
ORDER BY
    2 DESC
`)
	return items, err
}

func (r *StatsRepository) GetHighestRankedPersonByJob(userID string, job string) ([]types.HighestRated, error) {
	var persons []types.HighestRated
	err := r.db.Select(&persons, `
WITH person_ratings AS (
    SELECT
        p.name,
        p.id,
        COUNT(*) AS appearances,
        SUM(r.rating) AS total_rating
    FROM
        rating AS r
        INNER JOIN movie_person AS mp ON mp.movie_id = r.movie_id
            AND mp.job = $2
        INNER JOIN person AS p ON mp.person_id = p.id
    WHERE
        r.user_id = $1
    GROUP BY
        p.id
)
SELECT
    id,
    name,
    total_rating,
    (total_rating::float / appearances) * LOG(appearances) AS weighted_average_rating,
    appearances
FROM
    person_ratings
ORDER BY
    weighted_average_rating DESC
LIMIT 10
	`, userID, job)
	return persons, err
}
