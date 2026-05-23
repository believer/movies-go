package db

import (
	"github.com/jmoiron/sqlx"
)

type SettingsQuerier interface {
	GetWatchProviders(userID string) (string, error)
	UpdateWatchProviders(userID string, providers string) error
}

type SettingsRepository struct {
	db *sqlx.DB
}

func NewSettingsRepository(db *sqlx.DB) *SettingsRepository {
	return &SettingsRepository{db}
}

func (r *SettingsRepository) GetWatchProviders(userID string) (string, error) {
	var storedProviders string
	err := r.db.Get(&storedProviders, `SELECT watch_providers FROM "user" WHERE id = $1`, userID)
	return storedProviders, err
}

func (r *SettingsRepository) UpdateWatchProviders(userID string, providers string) error {
	_, err := r.db.Exec(`UPDATE "user" SET watch_providers = $1 WHERE id = $2`, providers, userID)
	return err
}
