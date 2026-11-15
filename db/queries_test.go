package db

import (
	"believer/movies/db/dbtest"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

var sqlStart = regexp.MustCompile(`(?i)^\s*(SELECT|INSERT|UPDATE|DELETE|WITH)\b`)

// TestQueriesPrepare has Postgres prepare every SQL literal in this package's
// source, so a syntax error or an unknown table or column fails here even when
// no other test runs the query. It needs no data: preparing checks the SQL
// against the schema without running it.
func TestQueriesPrepare(t *testing.T) {
	db := dbtest.Open(t)

	files, err := filepath.Glob("*.go")

	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	count := 0

	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)

		if err != nil {
			t.Fatal(err)
		}

		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)

			if !ok || lit.Kind != token.STRING {
				return true
			}

			query, err := strconv.Unquote(lit.Value)

			if err != nil || !sqlStart.MatchString(query) {
				return true
			}

			count++

			stmt, err := db.Preparex(query)

			if err != nil {
				t.Errorf("%s: %v", fset.Position(lit.Pos()), err)
				return true
			}

			stmt.Close()

			return true
		})
	}

	t.Logf("prepared %d queries", count)

	// Guards the finder itself: if it stopped matching, the test would pass
	// without preparing anything
	if count < 50 {
		t.Fatalf("found only %d queries, expected the repositories' queries", count)
	}
}
