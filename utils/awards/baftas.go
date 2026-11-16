package awards

import (
	"believer/movies/types"
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// AddBaftas adds the BAFTAs in baftas.csv for the movie with the IMDb ID id.
func AddBaftas(tx *sqlx.Tx, id string) error {
	if id == "" {
		return nil
	}

	data, err := files.ReadFile("baftas.csv")

	if err != nil {
		return err
	}

	csvReader := csv.NewReader(bytes.NewReader(data))
	headers, err := csvReader.Read()

	if err != nil {
		return fmt.Errorf("read baftas.csv: %w", err)
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
			return fmt.Errorf("read baftas.csv: %w", err)
		}

		imdbId := r[fields["FilmId"]]

		// Use to update one movie
		if imdbId != id {
			continue
		}

		year := r[fields["Year"]]
		category := r[fields["Category"]]
		title := r[fields["Film"]]
		winner := r[fields["Winner"]] == "TRUE"

		yearAsInt, err := strconv.Atoi(year)

		if err != nil {
			slog.Error("Could not parse BAFTA year", "year", year)
			continue
		}

		// We can only add where movie exists in database, otherwise
		// we get a violation on the foreign key to the movie table
		var movie types.Movie
		err = tx.Get(&movie, `
SELECT
    id,
    imdb_id
FROM
    movie
WHERE
    imdb_id = $1
    OR (LOWER(title) = LOWER($2)
        AND EXTRACT(year FROM release_date) = $3)
		`, imdbId, title, yearAsInt-1)

		if errors.Is(err, sql.ErrNoRows) {
			continue
		}

		if err != nil {
			return err
		}

		if movie.ImdbId == "" {
			continue
		}

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
				`, movie.ImdbId, n)

			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			if err != nil {
				return err
			}

			if person.Name == "" {
				continue
			}

			_, err = tx.Exec(`
				INSERT INTO award (name, imdb_id, winner, YEAR, person, person_id, type)
				    VALUES ($1, $2, $3, $4, $5, $6, 'bafta')
				ON CONFLICT (imdb_id, name, YEAR, person, detail)
				    DO UPDATE SET
				        winner = excluded.winner,
				        name = excluded.name,
				        person_id = excluded.person_id
			`, category, movie.ImdbId, winner, year, n, person.ID)

			if err != nil {
				return fmt.Errorf("add BAFTA %s %s to %s: %w", year, category, n, err)
			}

			slog.Debug("Added BAFTA", "category", category, "person", n)
		}
	}

	return nil
}
