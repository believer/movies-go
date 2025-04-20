-- Gets cast and crew for a movie.
-- The names, ids, and jobs are separated into arrays that are later zipped together in the backend.
-- name: cast-by-id
SELECT
    CASE mp.job
    WHEN 'cinematographer' THEN
        'Director of Photography'
    ELSE
        INITCAP(mp.job::text)
    END AS job,
    ARRAY_AGG(p.name ORDER BY num_movies DESC, p.popularity DESC) AS people_names,
    ARRAY_AGG(p.id ORDER BY num_movies DESC, p.popularity DESC) AS people_ids,
    CASE mp.job
    WHEN 'cast' THEN
        ARRAY_AGG(COALESCE(mp.character, '')
        ORDER BY num_movies DESC, p.popularity DESC)
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
    OR m.original_title ILIKE '%' || $1 || '%'
    OR se.name ILIKE '%' || $1 || '%'
ORDER BY
    m.release_date DESC;

-- name: feed-search-job
SELECT
    p.id,
    p.name,
    count(*)
FROM
    person p
    INNER JOIN movie_person mp ON mp.person_id = p.id
WHERE
    p."name" ILIKE '%' || $1 || '%'
    AND mp.job = $2
GROUP BY
    p.id
ORDER BY
    COUNT DESC
LIMIT 100;

-- name: feed-search-rating
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
    LEFT JOIN rating AS r ON r.movie_id = m.id
WHERE
    r.rating = $1
    AND r.user_id = $2
ORDER BY
    m.release_date DESC;

-- name: seen-by-user-id
SELECT
    id,
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
    get_person_role_with_seen_json (p.id, 'producer', $2) AS producer,
    get_person_role_with_seen_json (p.id, 'cinematographer', $2) AS cinematographer,
    get_person_role_with_seen_json (p.id, 'editor', $2) AS editor
FROM
    person AS p
WHERE
    p.id = $1;

-- name: awards-by-person-id
SELECT
    a.name AS "category",
    a.detail,
    a.winner,
    a.year,
    m.title,
    m.id AS "movie_id"
FROM
    award a
    INNER JOIN movie m ON m.imdb_id = a.imdb_id
WHERE
    person_id = $1
ORDER BY
    a.year DESC;

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

-- name: wilhelm-screams
SELECT
    count(*)
FROM
    seen s
    INNER JOIN movie m ON m.id = s.movie_id
WHERE
    user_id = $1
    AND m.wilhelm = TRUE;

