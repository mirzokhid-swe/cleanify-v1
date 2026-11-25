package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func testAuthCode(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func TestAuthCodeUse(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAuthCodeRepo(db)

	userID := insertTestUser(t, db, "Test auth code")
	code := testAuthCode(t)
	t.Cleanup(func() { db.Exec(`delete from auth_codes where code = $1`, code) })

	if err := repo.Create(userID, code, time.Now().Add(5*time.Minute)); err != nil {
		t.Fatalf("create: %v", err)
	}

	gotID, err := repo.Use(code)
	if err != nil {
		t.Fatalf("use: %v", err)
	}
	if gotID != userID {
		t.Fatalf("user id: got %d, want %d", gotID, userID)
	}

	// ikkinchi marta — ishlatilgan kod o'tmaydi
	if _, err := repo.Use(code); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("second use: got %v, want sql.ErrNoRows", err)
	}
}

func TestAuthCodeUseExpired(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAuthCodeRepo(db)

	userID := insertTestUser(t, db, "Test auth code expired")
	code := testAuthCode(t)
	t.Cleanup(func() { db.Exec(`delete from auth_codes where code = $1`, code) })

	if err := repo.Create(userID, code, time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := repo.Use(code); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expired use: got %v, want sql.ErrNoRows", err)
	}
}

func TestAuthCodeUseUnknown(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewAuthCodeRepo(db)

	if _, err := repo.Use("000000"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown use: got %v, want sql.ErrNoRows", err)
	}
}
