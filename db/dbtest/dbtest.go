// Package dbtest connects tests to a real Postgres database loaded with
// db/schema.sql. Tests using it are skipped with -short and fail when
// TEST_DATABASE_URL is unset.
package dbtest

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Open connects to TEST_DATABASE_URL and closes the connection when the test
// ends. Tests sharing the database must not run in parallel.
func Open(t *testing.T) *sqlx.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("database test skipped with -short")
	}

	url := os.Getenv("TEST_DATABASE_URL")

	if url == "" {
		t.Fatal("TEST_DATABASE_URL not set: run make test PKG=./<pkg> RUN=<Test>, or pass -short to skip database tests")
	}

	db, err := sqlx.Connect("postgres", url)

	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	return db
}

// Scratch creates an empty database on the TEST_DATABASE_URL server, for a
// test that needs a database without the schema, and drops it when the test
// ends.
func Scratch(t *testing.T) *sqlx.DB {
	t.Helper()

	server := Open(t)
	name := fmt.Sprintf("scratch_%d", time.Now().UnixNano())

	if _, err := server.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatalf("create scratch database: %v", err)
	}

	t.Cleanup(func() {
		if _, err := server.Exec("DROP DATABASE " + name + " WITH (FORCE)"); err != nil {
			t.Errorf("drop scratch database: %v", err)
		}
	})

	u, err := url.Parse(os.Getenv("TEST_DATABASE_URL"))

	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}

	u.Path = "/" + name
	db, err := sqlx.Connect("postgres", u.String())

	if err != nil {
		t.Fatalf("connect to scratch database: %v", err)
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
