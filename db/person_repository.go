package db

import (
	"believer/movies/types"
	"sort"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type Award string
type GroupedAwards map[string][]types.Award

const (
	AcademyAward Award = "academy-award"
	Bafta        Award = "bafta"
)

type PersonQuerier interface {
	GetPersonByID(id, userID string) (types.Person, error)
	GetGroupedAwards(id string, award Award) (GroupedAwards, int, []string, error)
}

type PersonRepository struct {
	db *sqlx.DB
}

func NewPersonRepository(db *sqlx.DB) *PersonRepository {
	return &PersonRepository{db}
}

func (r *PersonRepository) GetPersonByID(id, userID string) (types.Person, error) {
	var person types.Person
	err := r.db.Get(&person, personByIdQuery, id, userID)
	return person, err
}

func (r *PersonRepository) GetGroupedAwards(id string, award Award) (GroupedAwards, int, []string, error) {
	var awards types.Awards

	err := r.db.Select(&awards, awardsQuery, id, award)

	groupedAwards := make(GroupedAwards)

	wins := 0
	for _, award := range awards {
		if award.Winner {
			wins++
		}

		groupedAwards[award.Category] = append(groupedAwards[award.Category], award)
	}

	// Awards map is unsorted, create a sort order
	order := make([]string, 0, len(groupedAwards))
	for k := range groupedAwards {
		order = append(order, k)
	}

	sort.Strings(order)

	return groupedAwards, wins, order, err
}

// Queries
// =====================================================

const personByIdQuery = `
SELECT
    p.id,
    p.name,
    -- Function get_person_role_with_seen_json returns a JSON array of movies
    -- The function is defined in the database
    get_person_role_with_seen_json (p.id, 'director'::job, $2) AS director,
    get_person_role_with_seen_json (p.id, 'cast', $2) AS cast,
    get_person_role_with_seen_json (p.id, 'writer', $2) AS writer,
    get_person_role_with_seen_json (p.id, 'composer', $2) AS composer,
    get_person_role_with_seen_json (p.id, 'producer', $2) AS producer,
    get_person_role_with_seen_json (p.id, 'cinematographer', $2) AS cinematographer,
    get_person_role_with_seen_json (p.id, 'editor', $2) AS editor
FROM
    person AS p
WHERE
    p.id = $1
`

const awardsQuery = `
SELECT
    a.name AS "category",
    a.detail,
    a.winner,
    a.year,
    m.title,
    m.id AS "movie_id"
FROM
    award a
    INNER JOIN movie m ON m.imdb_id = a.imdb_id
WHERE
    person_id = $1
    AND type = $2
ORDER BY
    a.year DESC
`
