-- name: movies-by-number-of-awards
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
    count(*) = $2
ORDER BY
    m.release_date DESC;

