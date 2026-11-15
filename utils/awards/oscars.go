package awards

import (
	"believer/movies/types"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

// AddOscars adds the Academy Awards in oscars.csv for the movie with the
// IMDb ID id. The file is read from the working directory.
func AddOscars(tx *sqlx.Tx, id string) error {
	if id == "" {
		return nil
	}

	f, err := os.Open("oscars.csv")

	if err != nil {
		return err
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	headers, err := csvReader.Read()

	if err != nil {
		return fmt.Errorf("read oscars.csv: %w", err)
	}

	// Gather indexes to headers for easier mapping
	fields := make(map[string]int)
	for i, name := range headers {
		fields[name] = i
	}

	for {
		r, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read oscars.csv: %w", err)
		}

		year := r[fields["Year"]]
		category := r[fields["Category"]]
		imdbId := r[fields["FilmId"]]
		detail := r[fields["Detail"]]
		winner := r[fields["Winner"]] == "TRUE"

		// Use to update one movie
		if imdbId != id {
			continue
		}

		// Use to update a year
		// if year != "2025" {
		// 	continue
		// }

		slog.Debug("Found Academy Award", "movie", r[fields["Film"]])

		// We can only add where movie exists in database, otherwise
		// we get a violation on the foreign key to the movie table
		var movie types.Movie
		err = tx.Get(&movie, `SELECT
    id
FROM
    movie
WHERE
    imdb_id = $1`, imdbId)

		if errors.Is(err, sql.ErrNoRows) {
			continue
		}

		if err != nil {
			return err
		}

		switch category {
		case "Actor in a Leading Role",
			"Actor in a Supporting Role",
			"Actress in a Leading Role",
			"Actress in a Supporting Role",
			"Cinematography",
			"Directing",
			"Film Editing",
			"Music (Original Score)",
			"Writing (Adapted Screenplay)",
			"Writing (Original Screenplay)":
			nominees := strings.Split(r[fields["Nominees"]], ", ")

			if len(nominees) == 0 {
				continue
			}

			for _, n := range nominees {
				var person types.Person

				err = tx.Get(&person, `
					SELECT
					    p."name",
					    p.id
					FROM
					    movie m
					    INNER JOIN movie_person mp ON mp.movie_id = m.id
					    INNER JOIN person p ON p.id = mp.person_id
					WHERE
					    m.imdb_id = $1
					    AND p."name" ILIKE '%' || $2 || '%'
					`, imdbId, n)

				if errors.Is(err, sql.ErrNoRows) {
					continue
				}

				if err != nil {
					return err
				}

				_, err = tx.Exec(`
					INSERT INTO award (name, imdb_id, winner, YEAR, person, person_id, type)
					    VALUES ($1, $2, $3, $4, $5, $6, 'academy-award')
					ON CONFLICT (imdb_id, name, YEAR, person, detail)
					    DO UPDATE SET
					        winner = excluded.winner,
					        name = excluded.name,
					        person_id = excluded.person_id
				`, category, imdbId, winner, year, n, person.ID)

				if err != nil {
					return fmt.Errorf("add Academy Award %s %s to %s: %w", year, category, n, err)
				}

				slog.Debug("Added Academy Award", "category", category, "person", n)
			}
		case "Music (Original Song)":
			_, err = tx.Exec(`
				INSERT INTO award (name, imdb_id, winner, YEAR, detail, type)
				    VALUES ($1, $2, $3, $4, $5, 'academy-award')
				ON CONFLICT (imdb_id, name, YEAR, person, detail)
				    DO UPDATE SET
				        winner = excluded.winner
			`, category, imdbId, winner, year, detail)

			if err != nil {
				return fmt.Errorf("add Academy Award %s %s for %s: %w", year, category, detail, err)
			}

			slog.Debug("Added Academy Award", "category", category, "detail", detail)
		default:
			_, err = tx.Exec(`
				INSERT INTO award (name, imdb_id, winner, YEAR, type)
				    VALUES ($1, $2, $3, $4, 'academy-award')
				ON CONFLICT (imdb_id, name, YEAR, person, detail)
				    DO UPDATE SET
				        winner = excluded.winner
			`, category, imdbId, winner, year)

			if err != nil {
				return fmt.Errorf("add Academy Award %s %s: %w", year, category, err)
			}

			slog.Debug("Added Academy Award", "category", category, "imdb_id", imdbId)
		}
	}

	return nil
}
