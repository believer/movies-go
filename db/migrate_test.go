package db

import (
	"believer/movies/db/dbtest"
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadSQL runs a schema file the way psql would. The dumps set an empty
// search_path for the session, so the settings are reset afterwards.
func loadSQL(t *testing.T, db *sqlx.DB, file string) {
	t.Helper()

	sql, err := os.ReadFile(file)
	require.NoError(t, err)

	_, err = db.Exec(string(sql))
	require.NoError(t, err)

	_, err = db.Exec(`RESET ALL`)
	require.NoError(t, err)
}

func appliedVersions(t *testing.T, db *sqlx.DB) []int64 {
	t.Helper()

	var versions []int64
	err := db.Select(&versions, `SELECT version_id FROM goose_db_version WHERE is_applied ORDER BY id`)
	require.NoError(t, err)

	return versions
}

func TestMigrateEmptyDatabase(t *testing.T) {
	db := dbtest.Scratch(t)

	// One connection, so the queries below run on the one the baseline used
	db.SetMaxOpenConns(1)

	require.NoError(t, Migrate(context.Background(), db.DB))

	assert.Equal(t, []int64{0, 1}, appliedVersions(t, db)[:2])

	// Unqualified, so this fails if the baseline left an empty search_path
	var count int
	require.NoError(t, db.Get(&count, `SELECT count(*) FROM movie`))
	assert.Equal(t, 0, count)
}

// Production had the schema before it had migrations. Migrating it records
// the baseline without running it, and keeps its data.
func TestMigrateBaselinesExistingSchema(t *testing.T) {
	db := dbtest.Scratch(t)
	loadSQL(t, db, "migrations/baseline.sql")

	_, err := db.Exec(`INSERT INTO movie (title, imdb_id) VALUES ('Inception', 'tt1375666')`)
	require.NoError(t, err)

	require.NoError(t, Migrate(context.Background(), db.DB))

	assert.Equal(t, []int64{0, 1}, appliedVersions(t, db)[:2])

	var titles []string
	require.NoError(t, db.Select(&titles, `SELECT title FROM movie`))
	assert.Equal(t, []string{"Inception"}, titles)
}

// db/schema.sql, which tests and E2E load, records every migration, so the
// app starting against it has nothing to apply.
func TestMigrateSchemaFileIsUpToDate(t *testing.T) {
	db := dbtest.Scratch(t)
	loadSQL(t, db, "schema.sql")
	before := appliedVersions(t, db)

	require.NoError(t, Migrate(context.Background(), db.DB))

	assert.Equal(t, before, appliedVersions(t, db))
}
