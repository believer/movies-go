-- +goose Up
-- A user can watch a movie many times, and the movie counts once (GLOSSARY.md).
-- One row per user and movie the user has a Watch of.
CREATE VIEW watched_movie AS
SELECT DISTINCT user_id, movie_id
FROM seen;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.get_person_role_with_seen_json(pid integer, role public.job, uid integer) RETURNS json
    LANGUAGE sql STABLE
    AS $$
	SELECT
		ARRAY_TO_JSON(ARRAY (
				SELECT
					json_build_object('id', m.id, 'title', m.title, 'releaseDate', to_char(m.release_date, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), 'seen', w.movie_id IS NOT NULL, 'character', mp."character")
					FROM public.movie_person AS mp
					INNER JOIN public.movie AS m ON m.id = mp.movie_id
					LEFT JOIN public.watched_movie AS w ON w.movie_id = m.id
						AND w.user_id = uid
					WHERE
						mp.person_id = pid
						AND mp.job = ROLE
					ORDER BY
						m.release_date DESC))
$$;
-- +goose StatementEnd
