package db

import (
	"github.com/jmoiron/sqlx"
)

type UserAuth struct {
	PasswordHash string `db:"password_hash"`
	ID           string `db:"id"`
}

type AuthQuerier interface {
	GetUserForLogin(username string) (UserAuth, error)
	CreateUser(username string, passwordHash string) error
}

type AuthRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db}
}

func (r *AuthRepository) GetUserForLogin(username string) (UserAuth, error) {
	var user UserAuth
	err := r.db.Get(&user, "SELECT id, password_hash FROM public.user WHERE username = $1", username)
	return user, err
}

func (r *AuthRepository) CreateUser(username string, passwordHash string) error {
	_, err := r.db.Exec(`INSERT INTO "user" (username, password_hash) VALUES ($1, $2)`, username, passwordHash)
	return err
}
