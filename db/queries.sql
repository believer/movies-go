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
    user_id = 1
ORDER BY
    s.date DESC OFFSET $1
LIMIT 20;

-- name: movie-by-id
SELECT
    m.*,
    r.rating,
    ARRAY_AGG(g.name) AS genres
FROM
    movie AS m
    INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
    INNER JOIN genre AS g ON g.id = mg.genre_id
    LEFT JOIN rating AS r ON r.movie_id = m.id
        AND r.user_id = 1
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
    AND user_id = 1
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
    user_id = 1;

-- name: stats-most-watched-movies
SELECT
    COUNT(*) AS count,
    m.title AS name,
    m.id
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
WHERE
    user_id = 1
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
    user_id = 1
GROUP BY
    rating
ORDER BY
    rating;

-- name: stats-most-watched-by-job
SELECT
    COUNT(*) AS count,
    p.name,
    p.id
FROM
    seen AS s
    INNER JOIN movie_person AS mp ON mp.movie_id = s.movie_id
    INNER JOIN person AS p ON p.id = mp.person_id
WHERE
    user_id = 1
    AND mp.job = $1
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
    user_id = 1
    -- 2011 is where all the data that I hadn't tracked
    -- before I started ended up. So, there's a bunch of
    -- movies that year.
    AND EXTRACT(YEAR FROM date) > 2011
GROUP BY
    label
ORDER BY
    label;

