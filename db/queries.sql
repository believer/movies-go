-- Gets cast and crew for a movie.
-- The names, ids, and jobs are separated into arrays that are later zipped together in the backend.
-- name: cast-by-id
SELECT
    INITCAP(mp.job::text) AS job,
    ARRAY_AGG(p.name ORDER BY num_movies DESC) AS people_names,
    ARRAY_AGG(p.id ORDER BY num_movies DESC) AS people_ids,
    CASE mp.job
    WHEN 'cast' THEN
        ARRAY_AGG(COALESCE(mp.character, '')
        ORDER BY num_movies DESC)
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
    END;

-- Used for the start page feed. Returns the 20 most recently watched.
-- Implements an infinite scroll that loads the next 20 when the user scrolls to the bottom.
-- name: feed
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
LIMIT 20;

-- name: feed-search
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
    OR se.name ILIKE '%' || $1 || '%'
ORDER BY
    m.release_date DESC;

-- name: movie-by-id
SELECT
    m.id,
    m.title,
    m.release_date,
    m.runtime,
    m.imdb_id,
    m.overview,
    m.tagline,
    se.name AS "series",
    se.id AS "series_id",
    ms.number_in_series,
    r.rating,
    ARRAY_TO_JSON(ARRAY_AGG(json_build_object('name', g.name, 'id', g.id))) AS genres
FROM
    movie AS m
    INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
    INNER JOIN genre AS g ON g.id = mg.genre_id
    LEFT JOIN rating AS r ON r.movie_id = m.id
        AND r.user_id = $2
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
WHERE
    m.id = $1
GROUP BY
    1,
    r.rating,
    se.id,
    ms.number_in_series;

-- name: movie-by-name
SELECT
    m.id,
    m.title,
    m.release_date,
    m.runtime,
    m.imdb_id,
    m.overview,
    m.tagline,
    se.name AS "series",
    se.id AS "series_id",
    ms.number_in_series,
    r.rating,
    ARRAY_TO_JSON(ARRAY_AGG(json_build_object('name', g.name, 'id', g.id))) AS genres
FROM
    movie AS m
    INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
    INNER JOIN genre AS g ON g.id = mg.genre_id
    LEFT JOIN rating AS r ON r.movie_id = m.id
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
WHERE
    -- Slugify function is defined in the database
    slugify (m.title)
    ILIKE '%' || slugify ($1) || '%'
GROUP BY
    1,
    r.rating,
    se.id,
    ms.number_in_series;

-- name: seen-by-user-id
SELECT
    date at time zone 'UTC' at time zone 'Europe/Stockholm' AS date
FROM
    seen
WHERE
    movie_id = $1
    AND user_id = $2
ORDER BY
    date DESC;

-- name: person-by-id
SELECT
    p.id,
    p.name,
    -- Function get_person_role_json returns a JSON array of movies
    -- The function is defined in the database
    get_person_role_with_seen_json (p.id, 'director'::job, $2) AS director,
    get_person_role_with_seen_json (p.id, 'cast', $2) AS cast,
    get_person_role_with_seen_json (p.id, 'writer', $2) AS writer,
    get_person_role_with_seen_json (p.id, 'composer', $2) AS composer,
    get_person_role_with_seen_json (p.id, 'producer', $2) AS producer
FROM
    person AS p
WHERE
    p.id = $1;

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
        AND r.created_at >= date_trunc('year', CURRENT_DATE)
    LEFT JOIN seen s ON s.movie_id = r.movie_id
        AND s.user_id = $1
        AND s.date >= date_trunc('year', CURRENT_DATE)
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

-- name: movies-by-year
SELECT
    m.id,
    m.title,
    m.release_date,
    m.imdb_id,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie AS m
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
    date_part('year', release_date) = $2
ORDER BY
    release_date;

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

-- name: genres-by-id
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    m.imdb_id,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie_genre mg
    INNER JOIN movie m ON m.id = mg.movie_id
    LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            public.seen
        WHERE
            user_id = $2
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id
WHERE
    mg.genre_id = $1
ORDER BY
    m.release_date DESC;

-- name: genre-by-id
SELECT
    name
FROM
    genre
WHERE
    id = $1;

-- name: wilhelm-screams
SELECT
    count(*)
FROM
    seen s
    INNER JOIN movie m ON m.id = s.movie_id
WHERE
    user_id = $1
    AND m.wilhelm = TRUE;

