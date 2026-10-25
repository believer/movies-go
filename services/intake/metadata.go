package intake

import (
	"believer/movies/types"
	"database/sql"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type newPerson struct {
	ID             int            `db:"id"`
	Name           string         `db:"name"`
	Job            sql.NullString `db:"job"`
	Character      sql.NullString `db:"character"`
	Popularity     float64        `db:"popularity"`
	ProfilePicture sql.NullString `db:"profile_picture"`
	MovieId        int            `db:"movie_id"`
}

func addCast(tx *sqlx.Tx, movieCast types.MovieCreditsResponse, movieId int) error {
	var err error
	var castStructs []newPerson
	var crewStructs []newPerson

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

		castStructs = append(castStructs, newPerson{
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

		crewStructs = append(crewStructs, newPerson{
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
			slog.Error("Could not insert person", "err", err)
			return err
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
			slog.Error("Could not insert movie_person", "err", err)
			return err
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
			slog.Error("Could not insert crew", "err", err)
			return err
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
			slog.Error("Could not insert movie_person crew", "err", err)
			return err
		}
	}

	return nil
}

func personExists(arr []newPerson, id int, job any) (int, bool) {
	for i, person := range arr {
		if person.ID == id && person.Job.String == job {
			return i, true
		}
	}

	return 0, false
}

func addLanguages(tx *sqlx.Tx, id int, movie types.MovieDetailsResponse) error {
	type Language struct {
		ISO639      string `db:"iso_639_1"`
		EnglishName string `db:"english_name"`
		Name        string `db:"name"`
		MovieId     int    `db:"movie_id"`
	}

	languages := make([]Language, len(movie.SpokenLanguages))

	for i, l := range movie.SpokenLanguages {
		languages[i] = Language{
			ISO639:      l.ISO639,
			Name:        l.Name,
			EnglishName: l.EnglishName,
			MovieId:     id,
		}
	}

	if len(languages) > 0 {
		_, err := tx.NamedExec(
			`INSERT INTO
LANGUAGE (name, english_name, iso_639_1)
    VALUES (:name, :english_name, :iso_639_1)
ON CONFLICT
    DO NOTHING`, languages,
		)

		if err != nil {
			slog.Error("Could not insert language", "err", err)
			return err
		}

		_, err = tx.NamedExec(
			`INSERT INTO movie_language (movie_id, language_id)
    VALUES (:movie_id, (
            SELECT
                id
            FROM
                LANGUAGE
            WHERE
                name = :name))
ON CONFLICT
    DO NOTHING`, languages,
		)

		if err != nil {
			slog.Error("Could not insert movie_language", "err", err)
			return err
		}
	}

	return nil
}

func addGenres(tx *sqlx.Tx, id int, movie types.MovieDetailsResponse) error {
	type Genre struct {
		Name    string `db:"name"`
		MovieId int    `db:"movie_id"`
	}

	genres := make([]Genre, len(movie.Genres))

	for i, genre := range movie.Genres {
		genres[i] = Genre{
			Name:    genre.Name,
			MovieId: id,
		}
	}

	if len(genres) > 0 {
		_, err := tx.NamedExec(`INSERT INTO genre (name)
    VALUES (:name)
ON CONFLICT (name)
    DO NOTHING`, genres)

		if err != nil {
			slog.Error("Could not insert genre", "err", err)
			return err
		}

		_, err = tx.NamedExec(`INSERT INTO movie_genre (movie_id, genre_id)
    VALUES (:movie_id, (
            SELECT
                id
            FROM
                genre
            WHERE
                name = :name))
ON CONFLICT
    DO NOTHING`, genres)

		if err != nil {
			slog.Error("Could not insert movie_genre", "err", err)
			return err
		}
	}

	return nil
}

func addCountries(tx *sqlx.Tx, id int, movie types.MovieDetailsResponse) error {
	for _, c := range movie.ProductionCountries {
		_, err := tx.Exec(`
			INSERT INTO movie_country (movie_id, country_id)
			    VALUES ($1, $2)
			ON CONFLICT
			    DO NOTHING
    `, id, c.ID)
		if err != nil {
			slog.Error("Could not insert movie country", "err", err)
			return err
		}
	}
	return nil
}

func addProductionCompanies(tx *sqlx.Tx, id int, movie types.MovieDetailsResponse) error {
	for _, c := range movie.ProductionCompanies {
		var pcID int

		err := tx.Get(&pcID, `
			INSERT INTO production_company (tmdb_id, name, country)
			    VALUES ($1, $2, NULLIF ($3, ''))
			ON CONFLICT (tmdb_id, name)
			    DO UPDATE SET
			        name = EXCLUDED.name
			    RETURNING
			        id
		`, c.ID, c.Name, c.OriginCountry)

		if err != nil {
			slog.Error("Could not add company", "err", err)
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO movie_company (movie_id, company_id)
			    VALUES ($1, $2)
			ON CONFLICT
			    DO NOTHING
		`, id, pcID)

		if err != nil {
			slog.Error("Could not add movie company mapping", "err", err)
			return err
		}
	}

	slog.Info("Inserted production companies")
	return nil
}
