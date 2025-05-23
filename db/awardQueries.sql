-- name: movies-by-number-of-wins
SELECT
    m.id,
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    award a
    INNER JOIN movie m ON m.imdb_id = a.imdb_id
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
    winner = TRUE
GROUP BY
    a.imdb_id,
    m.id,
    s.id
HAVING
    count(DISTINCT a.name) = $2
ORDER BY
    m.release_date DESC;

-- name: movies-by-number-of-nominations
SELECT
    m.id,
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    award a
    INNER JOIN movie m ON m.imdb_id = a.imdb_id
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
GROUP BY
    a.imdb_id,
    m.id,
    s.id
HAVING
    count(DISTINCT a.name) = $2
ORDER BY
    m.release_date DESC;

-- name: stats-most-award-wins
WITH all_persons AS (
    SELECT DISTINCT ON (mp.person_id)
        mp.person_id
    FROM
        seen s
        INNER JOIN movie_person mp ON mp.movie_id = s.movie_id
    WHERE
        s.user_id = $1
)
SELECT
    count(*) FILTER (WHERE winner = TRUE) AS COUNT,
    a.person,
    a.person_id
FROM
    all_persons ap
    INNER JOIN award a ON ap.person_id = a.person_id
GROUP BY
    a.person_id,
    person
HAVING
    count(*) FILTER (WHERE winner = TRUE) > 0
ORDER BY
    COUNT DESC
LIMIT 1;

-- name: stats-most-award-nominations
WITH all_persons AS (
    SELECT DISTINCT ON (mp.person_id)
        mp.person_id
    FROM
        seen s
        INNER JOIN movie_person mp ON mp.movie_id = s.movie_id
    WHERE
        s.user_id = $1
)
SELECT
    count(*) AS COUNT,
    a.person,
    a.person_id
FROM
    all_persons ap
    INNER JOIN award a ON ap.person_id = a.person_id
GROUP BY
    a.person_id,
    person
HAVING
    count(*) > 0
ORDER BY
    COUNT DESC
LIMIT 1;

-- name: stats-top-awarded-movies
WITH movie_awards AS (
    SELECT
        m.id,
        m.title,
        COUNT(DISTINCT a.name) AS award_count
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
            movie_awards);

-- name: awards-by-year
SELECT
    m.id AS movie_id,
    m.title,
    COALESCE(JSONB_AGG(jsonb_build_object('person', a.person, 'person_id', a.person_id, 'winner', a.winner, 'category', a.name, 'detail', a.detail)
        ORDER BY a.winner DESC, a.name ASC), '[]'::jsonb) AS awards
FROM
    award AS a
    INNER JOIN movie AS m ON m.imdb_id = a.imdb_id
WHERE
    a.year = $1
GROUP BY
    m.id,
    m.title
ORDER BY
    m.title ASC
