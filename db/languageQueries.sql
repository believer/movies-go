-- name: movies-by-language-id
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    m.imdb_id,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie_language ml
    INNER JOIN movie m ON m.id = ml.movie_id
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
    ml.language_id = $1
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50;

-- name: language-by-id
SELECT
    id,
    english_name AS "name"
FROM
    "language"
WHERE
    id = $1;

