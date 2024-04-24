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
    s.date at time zone 'UTC' at time zone 'Europe/Stockholm' AS watched_at
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
WHERE
    user_id = $2
ORDER BY
    s.date DESC OFFSET $1
LIMIT 20;

-- name: movie-by-id
SELECT
    m.id,
    m.title,
    m.release_date,
    m.runtime,
    m.imdb_id,
    m.overview,
    m.tagline,
    r.rating,
    ARRAY_AGG(g.name) AS genres
FROM
    movie AS m
    INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
    INNER JOIN genre AS g ON g.id = mg.genre_id
    LEFT JOIN rating AS r ON r.movie_id = m.id
        AND r.user_id = $2
WHERE
    m.id = $1
GROUP BY
    1,
    r.rating;

-- name: movie-by-name
SELECT
    m.*,
    r.rating,
    ARRAY_AGG(g.name) AS genres
FROM
    movie AS m
    INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
    INNER JOIN genre AS g ON g.id = mg.genre_id
    INNER JOIN rating AS r ON r.movie_id = m.id
WHERE
    -- Slugify function is defined in the database
    slugify (m.title)
    ILIKE '%' || slugify ($1) || '%'
GROUP BY
    1,
    r.rating;

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
    get_person_role_json (p.id, 'director'::job) AS director,
    get_person_role_json (p.id, 'cast') AS cast,
    get_person_role_json (p.id, 'writer') AS writer,
    get_person_role_json (p.id, 'composer') AS composer,
    get_person_role_json (p.id, 'producer') AS producer
FROM
    person AS p
WHERE
    p.id = $1;

-- name: stats-data
SELECT
    COUNT(DISTINCT movie_id) AS unique_movies,
    COUNT(movie_id) seen_with_rewatches,
    SUM(m.runtime) AS total_runtime,
    MAX(m.imdb_rating) AS top_imdb_rating,
    (
        SELECT
            title
        FROM
            movie
        WHERE
            imdb_rating IS NOT NULL
        ORDER BY
            imdb_rating DESC
        LIMIT 1) AS top_imdb_title,
(
    SELECT
        id
    FROM
        movie
    WHERE
        imdb_rating IS NOT NULL
    ORDER BY
        imdb_rating DESC
    LIMIT 1) AS top_imdb_id
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
HAVING
    COUNT(*) > 1
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
WITH ratings AS (
    SELECT
        GENERATE_SERIES(1, 10) AS rating
)
SELECT
    ratings.rating AS label,
    COUNT(r.rating) AS value
FROM
    ratings
    LEFT JOIN rating AS r ON r.rating = ratings.rating
        AND r.user_id = $1
        AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM CURRENT_DATE)
GROUP BY
    ratings.rating,
    r.rating;

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
        user_id = $2) AS s
    INNER JOIN movie_person AS mp ON mp.movie_id = s.movie_id
    INNER JOIN person AS p ON p.id = mp.person_id
WHERE
    mp.job = $1
GROUP BY
    p.id
ORDER BY
    count DESC
LIMIT 10;

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
        generate_series(DATE_TRUNC('year', CURRENT_DATE), DATE_TRUNC('year', CURRENT_DATE) + INTERVAL '1 year' - INTERVAL '1 day', INTERVAL '1 month'))
SELECT
    TO_CHAR(months.month, 'Mon') AS label,
    COALESCE(count(seen.id), 0) AS value
FROM
    months
    LEFT JOIN seen ON DATE_TRUNC('month', seen.date) = months.month
        AND user_id = $1
WHERE
    EXTRACT(YEAR FROM seen.date) = EXTRACT(YEAR FROM CURRENT_DATE)
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
    rating AS r
    INNER JOIN movie AS m ON m.id = r.movie_id
WHERE
    EXTRACT(YEAR FROM r.created_at) = EXTRACT(YEAR FROM CURRENT_DATE)
    AND user_id = $1
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

