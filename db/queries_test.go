package db

import (
	"believer/movies/db/dbtest"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var sqlStart = regexp.MustCompile(`(?i)^[\s(]*(SELECT|INSERT|UPDATE|DELETE|WITH)\b`)

// sqlLiteral is a SQL string in this package's source.
type sqlLiteral struct {
	// Pos is the literal's file:line:column
	Pos string
	// Owner is the declaration holding the literal: a query constant's name
	// (genreMoviesQuery) or a function's (StatsRepository.GetRatings)
	Owner string
	Query string
}

// sqlLiterals returns every SQL string literal in this package's non-test
// source.
func sqlLiterals(t *testing.T) []sqlLiteral {
	t.Helper()

	files, err := filepath.Glob("*.go")

	if err != nil {
		t.Fatal(err)
	}

	files = slices.DeleteFunc(files, func(file string) bool { return strings.HasSuffix(file, "_test.go") })
	literals := sqlLiteralsIn(t, files...)

	// Guards the finder itself: if it stopped matching, the tests using it
	// would pass without checking anything
	if len(literals) < 50 {
		t.Fatalf("found only %d queries, expected the repositories' queries", len(literals))
	}

	return literals
}

// sqlLiteralsIn returns every SQL string literal in the given files.
func sqlLiteralsIn(t *testing.T, files ...string) []sqlLiteral {
	t.Helper()

	fset := token.NewFileSet()
	literals := []sqlLiteral{}

	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)

		if err != nil {
			t.Fatal(err)
		}

		for _, decl := range f.Decls {
			for _, owned := range ownedNodes(decl) {
				ast.Inspect(owned.node, func(n ast.Node) bool {
					lit, ok := n.(*ast.BasicLit)

					if !ok || lit.Kind != token.STRING {
						return true
					}

					query, err := strconv.Unquote(lit.Value)

					if err != nil || !sqlStart.MatchString(query) {
						return true
					}

					literals = append(literals, sqlLiteral{fset.Position(lit.Pos()).String(), owned.name, query})

					return true
				})
			}
		}
	}

	return literals
}

type ownedNode struct {
	name string
	node ast.Node
}

// ownedNodes splits a top-level declaration into the parts that own a query:
// a function, or each name in a const or var declaration.
func ownedNodes(decl ast.Decl) []ownedNode {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		name := d.Name.Name

		if d.Recv != nil && len(d.Recv.List) == 1 {
			recv := d.Recv.List[0].Type

			if star, ok := recv.(*ast.StarExpr); ok {
				recv = star.X
			}

			if ident, ok := recv.(*ast.Ident); ok {
				name = ident.Name + "." + name
			}
		}

		return []ownedNode{{name, d}}
	case *ast.GenDecl:
		nodes := []ownedNode{}

		for _, spec := range d.Specs {
			v, ok := spec.(*ast.ValueSpec)

			if !ok {
				continue
			}

			for i, name := range v.Names {
				if i < len(v.Values) {
					nodes = append(nodes, ownedNode{name.Name, v.Values[i]})
				}
			}
		}

		return nodes
	}

	return nil
}

// TestQueriesPrepare has Postgres prepare every SQL literal in this package's
// source, so a syntax error or an unknown table or column fails here even when
// no other test runs the query. It needs no data: preparing checks the SQL
// against the schema without running it.
//
// It prepares the benchmarks' queries too: they only run under make bench,
// so a schema change that breaks them would otherwise wait for the next
// performance check.
func TestQueriesPrepare(t *testing.T) {
	db := dbtest.Open(t)
	bench := sqlLiteralsIn(t, "bench_test.go")

	if len(bench) == 0 {
		t.Fatal("found no queries in bench_test.go")
	}

	literals := append(sqlLiterals(t), bench...)

	for _, lit := range literals {
		stmt, err := db.Preparex(lit.Query)

		if err != nil {
			t.Errorf("%s: %v", lit.Pos, err)
			continue
		}

		stmt.Close()
	}

	t.Logf("prepared %d queries", len(literals))
}
