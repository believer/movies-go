package dbtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaLoaded(t *testing.T) {
	db := Open(t)
	Truncate(t, db, "movie", "user")

	_, err := db.Exec(`INSERT INTO movie (title, imdb_id) VALUES ('Inception', 'tt1375666')`)
	assert.NoError(t, err)

	var count int
	err = db.Get(&count, `SELECT count(*) FROM movie`)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	Truncate(t, db, "movie")

	err = db.Get(&count, `SELECT count(*) FROM movie`)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
