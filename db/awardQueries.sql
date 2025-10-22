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

