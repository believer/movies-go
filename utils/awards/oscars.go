package awards

import (
	"believer/movies/db"
	"believer/movies/types"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func Add(id string) {
	f, err := os.Open("oscars.csv")

	if err != nil {
		panic(err)
	}

	defer func() {
		cerr := f.Close()

		if err != nil {
			err = cerr
		}
	}()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()

	if err != nil {
		panic(err)
	}

	// Gather indexes to headers for easier mapping
	fields := make(map[string]int)
	for i, name := range records[0] {
		fields[name] = i
	}

	tx := db.Client.MustBegin()

	for _, r := range records {
		year := r[fields["Year"]]
		category := r[fields["Category"]]
		imdbId := r[fields["FilmId"]]
		detail := r[fields["Detail"]]
		winner := r[fields["Winner"]] == "TRUE"

		// Use to update one movie
		if imdbId != id {
			continue
		}

		slog.Info("Found awards", "movie", r[fields["Film"]])

		// We can only add where movie exists in database, otherwise
		// we get a violation on the foreign key to the movie table
		var movie types.Movie
		err := tx.Get(&movie, `SELECT
    id
FROM
    movie
WHERE
    imdb_id = $1`, imdbId)

		if err != nil {
			continue
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

				if err != nil {
					continue
				}

				_, err = tx.Exec(`
					INSERT INTO award (name, imdb_id, winner, YEAR, person, person_id)
					    VALUES ($1, $2, $3, $4, $5, $6)
					ON CONFLICT (imdb_id, name, YEAR, person, detail)
					    DO UPDATE SET
					        winner = excluded.winner,
					        name = excluded.name,
					        person_id = excluded.person_id
				`, category, imdbId, winner, year, n, person.ID)

				if err != nil {
					fmt.Printf("Person Err %s %s %t %s %s %d\n", category, imdbId, winner, year, n, person.ID)
					panic(err)
				}

				fmt.Printf("Person %s\n", n)
			}
		case "Music (Original Song)":
			_, err = tx.Exec(`
				INSERT INTO award (name, imdb_id, winner, YEAR, detail)
				    VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (imdb_id, name, YEAR, person, detail)
				    DO UPDATE SET
				        winner = excluded.winner
			`, category, imdbId, winner, year, detail)

			if err != nil {
				fmt.Printf("Music Err %s %s %t %s %s\n", category, imdbId, winner, year, detail)
				panic(err)
			}

			fmt.Printf("Music %s\n", detail)
		default:
			_, err = tx.Exec(`
				INSERT INTO award (name, imdb_id, winner, YEAR)
				    VALUES ($1, $2, $3, $4)
				ON CONFLICT (imdb_id, name, YEAR, person, detail)
				    DO UPDATE SET
				        winner = excluded.winner
			`, category, imdbId, winner, year)

			if err != nil {
				fmt.Printf("Other Err %s %s %t %s\n", category, imdbId, winner, year)
				panic(err)
			}

			fmt.Printf("Other %s %s\n", category, imdbId)
		}
	}

	err = tx.Commit()

	if err != nil {
		err = tx.Rollback()

		panic(err)
	}
}
