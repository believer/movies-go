-- name: review-by-movie-id
SELECT
    id,
    content,
    private
FROM
    review
WHERE
    movie_id = $1
    AND user_id = $2
    AND private IS FALSE;

