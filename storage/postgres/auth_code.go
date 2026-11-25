package postgres

import (
	"bw-erp/storage/repo"
	"time"

	"github.com/jmoiron/sqlx"
)

type authCodeRepo struct {
	db *sqlx.DB
}

func NewAuthCodeRepo(db *sqlx.DB) repo.AuthCodeI {
	return &authCodeRepo{db: db}
}

func (stg authCodeRepo) Create(userID int64, code string, expiresAt time.Time) error {
	_, err := stg.db.Exec(`insert into auth_codes (user_id, code, expires_at) values ($1, $2, $3)`, userID, code, expiresAt)
	return err
}

func (stg authCodeRepo) Use(code string) (int64, error) {
	var userID int64
	err := stg.db.Get(&userID, `
		update auth_codes
		set used_at = now()
		where code = $1 and used_at is null and expires_at > now()
		returning user_id`, code)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
