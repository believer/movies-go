-- name: stats-data
SELECT
    COUNT(DISTINCT movie_id) AS unique_movies,
    COUNT(movie_id) seen_with_rewatches,
    COALESCE(SUM(m.runtime), 0) AS total_runtime
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
WHERE
    user_id = $1;

-- name: stats-most-watched-movies
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
LIMIT 20;

-- name: stats-ratings
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
    rating;

-- name: stats-ratings-this-year
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
    rs.rating_value;

-- name: stats-most-watched-by-job
SELECT
    COUNT(*) AS count,
    p.name,
    p.id
FROM ( SELECT DISTINCT ON (movie_id)
        movie_id
    FROM
        seen
    WHERE
        user_id = $2
        AND ($3 = 'All'
            OR EXTRACT(YEAR FROM date) = $3::int)) AS s
    INNER JOIN movie_person AS mp ON mp.movie_id = s.movie_id
    INNER JOIN person AS p ON p.id = mp.person_id
WHERE
    mp.job = $1
GROUP BY
    p.id
ORDER BY
    count DESC
LIMIT 10;

-- name: total-watched-by-job-and-year
SELECT
    COUNT(*) AS count
FROM
    seen s
    INNER JOIN movie_person mp ON mp.movie_id = s.movie_id
WHERE
    user_id = $1
    AND mp.job = $2
    AND ($3 = 'All'
        OR EXTRACT(YEAR FROM date) = $3::int);

-- name: stats-watched-by-year
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
    label;

-- name: stats-watched-this-year-by-month
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
    months.month;

-- name: stats-best-of-the-year
SELECT
    m.id,
    m.title,
    r.rating
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
    INNER JOIN rating AS r ON m.id = r.movie_id
WHERE
    EXTRACT(YEAR FROM s.date) = EXTRACT(YEAR FROM CURRENT_DATE)
    AND s.user_id = $1
ORDER BY
    rating DESC
LIMIT 1;

-- name: stats-movies-by-year
SELECT
    EXTRACT(YEAR FROM release_date) AS label,
    COUNT(*) AS value
FROM
    rating AS r
    INNER JOIN movie AS m ON m.id = r.movie_id
WHERE
    r.user_id = $1
GROUP BY
    label
ORDER BY
    label DESC;

-- name: shortest-and-longest-movie
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
    LIMIT 1);

-- name: stats-genres
SELECT
    g.id,
    g.name,
    COUNT(DISTINCT s.movie_id) AS count
FROM
    seen s
    INNER JOIN movie_genre mg ON mg.movie_id = s.movie_id
    INNER JOIN genre g ON mg.genre_id = g.id
WHERE
    s.user_id = $1
GROUP BY
    g.id
ORDER BY
    count DESC
LIMIT 10;
