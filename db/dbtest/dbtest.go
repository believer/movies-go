// Package dbtest connects tests to a real Postgres database loaded with
// db/schema.sql. Tests using it are skipped when TEST_DATABASE_URL is unset.
package dbtest

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Open connects to TEST_DATABASE_URL and closes the connection when the test
// ends. Tests sharing the database must not run in parallel.
func Open(t *testing.T) *sqlx.DB {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")

	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := sqlx.Connect("postgres", url)

	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	return db
}

// Truncate empties the given tables and resets their sequences.
func Truncate(t *testing.T, db *sqlx.DB, tables ...string) {
	t.Helper()

	quoted := make([]string, len(tables))

	for i, table := range tables {
		quoted[i] = pq.QuoteIdentifier(table)
	}

	_, err := db.Exec("TRUNCATE " + strings.Join(quoted, ", ") + " RESTART IDENTITY CASCADE")

	if err != nil {
		t.Fatalf("truncate %v: %v", tables, err)
	}
}
