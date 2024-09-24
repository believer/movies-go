-- name: series-by-id
SELECT
    *
FROM
    series
WHERE
    id = $1;

-- name: series-movies-by-id
WITH RECURSIVE series_hierarchy AS (
    -- Base case: Start with the parent series
    SELECT
        s.id AS series_id,
        s.name AS series_name,
        s.parent_series
    FROM
        series s
    WHERE
        s.id = $1
    UNION ALL
    -- Recursive case: Find all child series
    SELECT
        child.id AS series_id,
        child.name AS series_name,
        child.parent_series
    FROM
        series child
        INNER JOIN series_hierarchy sh ON child.parent_series = sh.series_id
)
SELECT
    sh.series_id AS "id",
    sh.series_name AS "name",
    ARRAY_TO_JSON(COALESCE(array_agg(json_build_object('id', m.id, 'title', m.title, 'release_date', to_char(m.release_date, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), 'number_in_series', ms.number_in_series, 'seen', (s.id IS NOT NULL))
            ORDER BY ms.number_in_series ASC) FILTER (WHERE m.id IS NOT NULL), '{}'::json[])) AS movies
FROM
    series_hierarchy sh
    LEFT JOIN movie_series ms ON ms.series_id = sh.series_id
    LEFT JOIN movie m ON m.id = ms.movie_id
    LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            seen
        WHERE
            user_id = $2
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id
GROUP BY
    sh.series_id,
    sh.series_name;

