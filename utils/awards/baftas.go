package awards

import (
	"believer/movies/types"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

func AddBaftas(tx *sqlx.Tx, id string) {
	f, err := os.Open("baftas.csv")

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

	for _, r := range records {
		imdbId := r[fields["FilmId"]]

		// Use to update one movie
		if imdbId != id {
			continue
		}

		year := r[fields["Year"]]
		category := r[fields["Category"]]
		title := r[fields["Film"]]
		winner := r[fields["Winner"]] == "TRUE"

		// if category == "EE Rising Star" {
		// 	var person types.Person
		// 	nominees := strings.Split(r[fields["Nominees"]], ", ")
		//
		// 	if len(nominees) == 0 {
		// 		continue
		// 	}
		//
		// 	n := nominees[0]
		//
		// 	err = tx.Get(&person, `
		// 		SELECT
		// 		    p."name",
		// 		    p.id
		// 		FROM
		// 		    person p
		// 		WHERE
		// 		    p."name" ILIKE '%' || $1 || '%'
		// 		`, n)
		//
		// 	if err != nil {
		// 		continue
		// 	}
		//
		// 	_, err = tx.Exec(`
		// 		INSERT INTO award (name, winner, YEAR, person, person_id, type)
		// 		    VALUES ($1, $2, $3, $4, $5, 'bafta')
		// 		ON CONFLICT (imdb_id, name, YEAR, person, detail)
		// 		    DO UPDATE SET
		// 		        winner = excluded.winner,
		// 		        name = excluded.name,
		// 		        person_id = excluded.person_id
		// 	`, category, winner, year, n, person.ID)
		//
		// 	if err != nil {
		// 		continue
		// 	}
		//
		// 	fmt.Printf("Rising Star %s\n", n)
		//
		// 	continue
		// }

		yearAsInt, err := strconv.Atoi(year)

		if err != nil {
			fmt.Printf("[BAFTA] Year error: %s", year)
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

		if err != nil || movie.ImdbId == "" {
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

			if err != nil || person.Name == "" {
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
				fmt.Printf("Person Err %s %s %t %s %s %d\n", category, imdbId, winner, year, n, person.ID)
				continue
			}

			fmt.Printf("Person %s\n", n)
		}
	}
}
