package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// migrations holds the SQL migrations, named NNNNN_name.sql, and baseline.sql,
// which goose skips because it has no version.
//
//go:embed migrations
var migrations embed.FS

// baseline is migration 1: the schema production had before there were
// migrations. A database that already has it, like production, records the
// migration as applied without running it.
var baseline = goose.NewGoMigration(1, &goose.GoFunc{RunTx: applyBaseline}, nil)

func applyBaseline(ctx context.Context, tx *sql.Tx) error {
	var exists bool

	err := tx.QueryRowContext(ctx, `SELECT to_regclass('public.movie') IS NOT NULL`).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	schema, err := migrations.ReadFile("migrations/baseline.sql")

	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, string(schema)); err != nil {
		return err
	}

	// The dump sets an empty search_path and other settings for the session.
	// Reset them, so goose can record the version and the pooled connection
	// works as before.
	_, err = tx.ExecContext(ctx, `RESET ALL`)

	return err
}

// Migrate applies the migrations db is missing. A Postgres advisory lock
// keeps two machines starting at once from running them both.
func Migrate(ctx context.Context, db *sql.DB) error {
	dir, err := fs.Sub(migrations, "migrations")

	if err != nil {
		return err
	}

	locker, err := lock.NewPostgresSessionLocker()

	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		dir,
		goose.WithGoMigrations(baseline),
		goose.WithDisableGlobalRegistry(true),
		goose.WithSessionLocker(locker),
	)

	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)

	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	for _, result := range results {
		slog.Info("Applied migration", "version", result.Source.Version, "duration", result.Duration)
	}

	return nil
}
