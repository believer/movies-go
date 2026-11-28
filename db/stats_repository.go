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
	GetHotTakes(userID string) (types.HotTakes, error)
	GetIMDbComparison(userID string) (types.IMDbComparison, error)
	GetStreaks(userID string) (types.Streaks, error)
	GetActorDirectorPairs(userID string) ([]types.PersonPair, error)
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
    latest_rating
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
SELECT
    COUNT(*)
FROM
    watched_movie w
    INNER JOIN movie m ON m.id = w.movie_id
        AND m.wilhelm = TRUE
WHERE
    w.user_id = $1
	`, userID)
	return wilhelms, err
}

func (r *StatsRepository) GetRatingsThisYear(userID string, yearTime string) ([]graph.GraphData, error) {
	var data []graph.GraphData
	err := r.db.Select(&data, `
WITH rating_series AS (
    SELECT
        generate_series(1, 10) AS rating_value
),
-- Each movie counts once, however many times it was watched in the year
watched_this_year AS (
    SELECT DISTINCT
        movie_id
    FROM
        seen
    WHERE
        user_id = $1
        AND EXTRACT(YEAR FROM date) = EXTRACT(YEAR FROM $2::date)
)
SELECT
    rs.rating_value AS label,
    COUNT(w.movie_id) AS value
FROM
    rating_series rs
    LEFT JOIN latest_rating lr ON lr.rating = rs.rating_value
        AND lr.user_id = $1
    LEFT JOIN watched_this_year w ON w.movie_id = lr.movie_id
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
SELECT DISTINCT
    m.id,
    m.title,
    rr.rating,
    s.date AS "watched_at"
FROM
    seen s
    INNER JOIN movie m ON m.id = s.movie_id
    INNER JOIN latest_rating rr ON rr.movie_id = s.movie_id
        AND rr.user_id = s.user_id
WHERE
    EXTRACT('YEAR' FROM s.date) = $1
    AND s.user_id = $2
ORDER BY
    rr.rating DESC,
    s.date ASC
	`, year, userID)
	return movies, err
}

func (r *StatsRepository) GetBestOfTheYear(userID string, year string) ([]types.ListItem, error) {
	var movies []types.ListItem
	err := r.db.Select(&movies, `
WITH watched_once AS (
    SELECT
        movie_id,
        MIN(date) AS date
    FROM
        seen
    WHERE
        user_id = $1
    GROUP BY
        movie_id
    HAVING
        -- Watched exactly once, and that watch is in the given year
        COUNT(*) = 1
        AND MIN(date) >= make_date($2, 1, 1)
        AND MIN(date) < make_date($2 + 1, 1, 1)
)
SELECT
    m.title AS "name",
    m.id AS "id",
    lr.rating AS "count"
FROM
    watched_once wo
    INNER JOIN latest_rating lr ON lr.movie_id = wo.movie_id
        AND lr.user_id = $1
    INNER JOIN movie m ON m.id = wo.movie_id
ORDER BY
    lr.rating DESC,
    wo.date ASC,
    m.id ASC
LIMIT 10
	`, userID, year)
	return movies, err
}

func (r *StatsRepository) GetWilhelmMovies(userID string, offset int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.release_date,
    TRUE AS "seen"
FROM
    watched_movie w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    w.user_id = $1
    AND m.wilhelm = TRUE
ORDER BY
    m.release_date DESC,
    m.id OFFSET $2
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
        mp.person_id,
        COUNT(*) AS appearances,
        SUM(r.rating) AS total_rating
    FROM
        latest_rating AS r
        INNER JOIN movie_person AS mp ON mp.movie_id = r.movie_id
            AND mp.job = $2
    WHERE
        r.user_id = $1
    GROUP BY
        mp.person_id
),
-- Ranks before joining person, so only the top 10 people get their names
-- looked up rather than every person the user rated a movie of
top_rated AS (
    SELECT
        person_id,
        total_rating,
        (total_rating::float / appearances) * LOG(appearances) AS weighted_average_rating,
        appearances
    FROM
        person_ratings
    ORDER BY
        weighted_average_rating DESC,
        person_id
    LIMIT 10
)
SELECT
    p.id,
    p.name,
    t.total_rating,
    t.weighted_average_rating,
    t.appearances
FROM
    top_rated AS t
    INNER JOIN person AS p ON p.id = t.person_id
ORDER BY
    t.weighted_average_rating DESC,
    p.id
	`, userID, job)
	return persons, err
}

// hotTakesQuery returns the user's 10 movies rated furthest from IMDb on one
// side of it: $2 is 1 for above and -1 for below.
const hotTakesQuery = `
SELECT
    m.id,
    m.title,
    lr.rating,
    i.imdb_rating::float8 AS imdb_rating,
    (lr.rating - i.imdb_rating)::float8 AS difference
FROM
    latest_rating AS lr
    INNER JOIN movie AS m ON m.id = lr.movie_id
    -- Checked before the cast, which a malformed rating would fail
    CROSS JOIN LATERAL (
        SELECT
            CASE WHEN m.imdb_rating ~ '^[0-9]+(\.[0-9]+)?$' THEN
                m.imdb_rating::numeric
            END AS imdb_rating) AS i
WHERE
    lr.user_id = $1
    -- IMDb's 0 means the movie has no rating there
    AND i.imdb_rating > 0
    AND SIGN(lr.rating - i.imdb_rating) = $2
ORDER BY
    ABS(lr.rating - i.imdb_rating) DESC,
    m.id
LIMIT 10
`

func (r *StatsRepository) GetHotTakes(userID string) (types.HotTakes, error) {
	var takes types.HotTakes

	if err := r.db.Select(&takes.Above, hotTakesQuery, userID, 1); err != nil {
		return takes, err
	}

	err := r.db.Select(&takes.Below, hotTakesQuery, userID, -1)
	return takes, err
}

func (r *StatsRepository) GetIMDbComparison(userID string) (types.IMDbComparison, error) {
	var comparison types.IMDbComparison
	err := r.db.Get(&comparison, `
SELECT
    COUNT(*) AS movies,
    COALESCE(AVG(lr.rating - i.imdb_rating), 0)::float8 AS average_difference
FROM
    latest_rating AS lr
    INNER JOIN movie AS m ON m.id = lr.movie_id
    -- Checked before the cast, which a malformed rating would fail
    CROSS JOIN LATERAL (
        SELECT
            CASE WHEN m.imdb_rating ~ '^[0-9]+(\.[0-9]+)?$' THEN
                m.imdb_rating::numeric
            END AS imdb_rating) AS i
WHERE
    lr.user_id = $1
    -- IMDb's 0 means the movie has no rating there
    AND i.imdb_rating > 0
	`, userID)
	return comparison, err
}

func (r *StatsRepository) GetStreaks(userID string) (types.Streaks, error) {
	var streaks types.Streaks
	err := r.db.Get(&streaks, `
WITH watch_days AS (
    SELECT
        date::date AS day,
        COUNT(*) AS watches,
        -- Numbers the days in order, so the earliest record wins a tie
        ROW_NUMBER() OVER (ORDER BY date::date) AS day_id
    FROM
        seen
    WHERE
        user_id = $1
        -- Watches dated 2011 are a bulk import of the years before
        -- tracking started
        AND date >= '2012-01-01'
    GROUP BY
        date::date
),
streak AS (
    SELECT
        MIN(day) AS start_day,
        MAX(day) AS end_day,
        COUNT(*) AS days,
        MIN(day_id) AS streak_id
    FROM
        watch_days
    GROUP BY
        -- Is the same for every day of a run of consecutive days
        day - day_id::int
    ORDER BY
        days DESC,
        streak_id
    LIMIT 1
),
gap AS (
    SELECT
        prev.day AS start_day,
        next.day AS end_day,
        next.day - prev.day - 1 AS days
    FROM
        watch_days AS prev
        INNER JOIN watch_days AS next ON next.day_id = prev.day_id + 1
    ORDER BY
        days DESC,
        prev.day_id
    LIMIT 1
),
busiest AS (
    SELECT
        day,
        watches
    FROM
        watch_days
    ORDER BY
        watches DESC,
        day_id
    LIMIT 1
)
SELECT
    COALESCE(s.days, 0) AS streak_days,
    s.start_day AS streak_start,
    s.end_day AS streak_end,
    COALESCE(g.days, 0) AS break_days,
    g.start_day AS break_start,
    g.end_day AS break_end,
    COALESCE(b.watches, 0) AS busiest_day_watches,
    b.day AS busiest_day,
    (
        SELECT
            COUNT(*)
        FROM
            watch_days
        WHERE
            watches > 1) AS binge_days
FROM (
    SELECT) AS one
    LEFT JOIN streak AS s ON TRUE
    LEFT JOIN gap AS g ON TRUE
    LEFT JOIN busiest AS b ON TRUE
	`, userID)
	return streaks, err
}

func (r *StatsRepository) GetActorDirectorPairs(userID string) ([]types.PersonPair, error) {
	var pairs []types.PersonPair
	err := r.db.Select(&pairs, `
WITH billed AS (
    SELECT
        mp.movie_id,
        mp.person_id,
        -- Movie intake adds the cast in billing order
        ROW_NUMBER() OVER (PARTITION BY mp.movie_id ORDER BY mp.id) AS billing
    FROM
        watched_movie AS w
        INNER JOIN movie_person AS mp ON mp.movie_id = w.movie_id
            AND mp.job = 'cast'
    WHERE
        w.user_id = $1
),
-- Ranks before joining person, so only the top 10 pairs get their names
-- looked up
pairs AS (
    SELECT
        c.person_id AS actor_id,
        d.person_id AS director_id,
        COUNT(*) AS movies
    FROM
        billed AS c
        INNER JOIN movie_person AS d ON d.movie_id = c.movie_id
            AND d.job = 'director'
            -- An actor directing themselves isn't a pair
            AND d.person_id <> c.person_id
    WHERE
        -- The leads, not every bit part
        c.billing <= 5
    GROUP BY
        c.person_id,
        d.person_id
    HAVING
        COUNT(*) > 1
    ORDER BY
        movies DESC,
        actor_id,
        director_id
    LIMIT 10
)
SELECT
    p.actor_id,
    a.name AS actor_name,
    p.director_id,
    d.name AS director_name,
    p.movies
FROM
    pairs AS p
    INNER JOIN person AS a ON a.id = p.actor_id
    INNER JOIN person AS d ON d.id = p.director_id
ORDER BY
    p.movies DESC,
    p.actor_id,
    p.director_id
	`, userID)
	return pairs, err
}
