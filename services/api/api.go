package api

import (
	"believer/movies/db"
	"believer/movies/services/tmdb"
	"believer/movies/types"
	"database/sql"
	"log/slog"
)

type Api struct{}

func New() *Api {
	return &Api{}
}

func (a *Api) AddMovie(imdbId string, hasWilhelmScream bool) (types.MovieDetailsResponse, int, error) {
	var id int

	tmdbApi := tmdb.New(imdbId)
	movie, err := tmdbApi.Movie()

	if err != nil {
		return types.MovieDetailsResponse{}, 0, err
	}

	tx := db.Client.MustBegin()

	err = tx.Get(
		&id,
		`
INSERT INTO movie (title, runtime, release_date, imdb_id, overview, poster, tagline, tmdb_id, wilhelm)
    VALUES ($1, $2, NULLIF ($3, '')::date, $4, $5, $6, $7, $8, $9)
ON CONFLICT (imdb_id)
    DO UPDATE SET
        title = excluded.title,
        runtime = excluded.runtime,
        release_date = excluded.release_date,
        imdb_id = excluded.imdb_id,
        overview = excluded.overview,
        poster = excluded.poster,
        tagline = excluded.tagline,
        tmdb_id = excluded.tmdb_id
    RETURNING
        id
		`,
		movie.Title,
		movie.Runtime,
		movie.ReleaseDate,
		movie.ImdbId,
		movie.Overview,
		movie.Poster,
		movie.Tagline,
		movie.TmdbId,
		hasWilhelmScream,
	)

	if err != nil {
		return movie, 0, err
	}

	err = tx.Commit()

	if err != nil {
		slog.Error("Could not commit movie")
		err = tx.Rollback()

		return movie, 0, err
	}

	a.AddCast(imdbId, id)

	return movie, id, nil
}

type NewPerson struct {
	ID             int            `db:"id"`
	Name           string         `db:"name"`
	Job            sql.NullString `db:"job"`
	Character      sql.NullString `db:"character"`
	Popularity     float64        `db:"popularity"`
	ProfilePicture sql.NullString `db:"profile_picture"`
	MovieId        int            `db:"movie_id"`
}

func (a *Api) AddCast(imdbId string, movieId int) {
	tmdbApi := tmdb.New(imdbId)
	movieCast, err := tmdbApi.Credits()

	if err != nil {
		slog.Error("Could not get TMDb credits")
		return
	}

	var castStructs []NewPerson
	var crewStructs []NewPerson

	tx := db.Client.MustBegin()

	// Insert cast
	for _, cast := range movieCast.Cast {
		var pfp sql.NullString
		var char sql.NullString

		if cast.ProfilePath == nil {
			pfp = sql.NullString{String: "", Valid: false}
		} else {
			pfp = sql.NullString{String: *cast.ProfilePath, Valid: true}
		}

		if cast.Character == nil {
			char = sql.NullString{String: "", Valid: false}
		} else {
			char = sql.NullString{String: *cast.Character, Valid: true}
		}

		personIndex, exists := personExists(castStructs, cast.ID, "cast")

		if exists {
			castStructs[personIndex].Name = cast.Name
			castStructs[personIndex].Popularity = cast.Popularity
			castStructs[personIndex].Character = char
			castStructs[personIndex].ProfilePicture = pfp

			continue
		}

		castStructs = append(castStructs, NewPerson{
			ID:             cast.ID,
			Name:           cast.Name,
			Popularity:     cast.Popularity,
			Job:            sql.NullString{String: "cast", Valid: true},
			Character:      char,
			ProfilePicture: pfp,
			MovieId:        movieId,
		})
	}

	// Crew
	for _, crew := range movieCast.Crew {
		department := crew.Department

		if department != "Directing" && department != "Writing" && department != "Production" && department != "Sound" && department != "Camera" && department != "Editing" {
			continue
		}

		var pfp sql.NullString

		if crew.ProfilePath == nil {
			pfp = sql.NullString{String: "", Valid: false}
		} else {
			pfp = sql.NullString{String: *crew.ProfilePath, Valid: true}
		}

		if crew.Job == nil {
			continue
		}

		job := *crew.Job

		switch job {
		case "Screenplay", "Writer", "Novel":
			job = "writer"
		case "Original Music Composer":
			job = "composer"
		case "Producer", "Associate Producer", "Executive Producer":
			job = "producer"
		case "Director":
			job = "director"
		case "Director of Photography":
			job = "cinematographer"
		case "Editor":
			job = "editor"
		default:
			continue
		}

		jobStr := sql.NullString{String: job, Valid: true}
		personIndex, exists := personExists(crewStructs, crew.Id, job)

		if exists {
			crewStructs[personIndex].Name = crew.Name
			crewStructs[personIndex].Popularity = crew.Popularity
			crewStructs[personIndex].ProfilePicture = pfp

			continue
		}

		crewStructs = append(crewStructs, NewPerson{
			ID:             crew.Id,
			Name:           crew.Name,
			Popularity:     crew.Popularity,
			Job:            jobStr,
			ProfilePicture: pfp,
			MovieId:        movieId,
		})
	}

	if len(castStructs) > 0 {
		_, err = tx.NamedExec(`
	INSERT INTO person (name, original_id, popularity, profile_picture)
	    VALUES (:name, :id, :popularity, :profile_picture)
	ON CONFLICT
	    DO NOTHING
	`, castStructs)

		if err != nil {
			slog.Error("Could not insert person")
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job, character)
	    VALUES (:movie_id, (
	            SELECT
	                id
	            FROM
	                person
	            WHERE
	                original_id = :id), 'cast', :character)
	ON CONFLICT (movie_id,
	    person_id,
	    job)
	    DO UPDATE SET
	        character = excluded.character
	`, castStructs)

		if err != nil {
			slog.Error("Could not insert movie_person")
		}
	}

	slog.Debug("Cast inserted")

	if len(crewStructs) > 0 {
		_, err = tx.NamedExec(`
	INSERT INTO person (name, original_id, popularity, profile_picture)
	    VALUES (:name, :id, :popularity, :profile_picture)
	ON CONFLICT
	    DO NOTHING
	`, crewStructs)

		if err != nil {
			slog.Error("Could not insert crew")
		}

		_, err = tx.NamedExec(`
	INSERT INTO movie_person (movie_id, person_id, job)
	    VALUES (:movie_id, (
	            SELECT
	                id
	            FROM
	                person
	            WHERE
	                original_id = :id), :job)
	ON CONFLICT (movie_id,
	    person_id,
	    job)
	    DO UPDATE SET
	        job = excluded.job
	`, crewStructs)

		if err != nil {
			slog.Error("Could not insert movie_person crew")
		}
	}

	err = tx.Commit()

	if err != nil {
		slog.Error("Could not commit cast")
		err = tx.Rollback()
	}
}

func personExists(arr []NewPerson, id int, job any) (int, bool) {
	for i, person := range arr {
		if person.ID == id && person.Job.String == job {
			return i, true
		}
	}

	return 0, false
}
