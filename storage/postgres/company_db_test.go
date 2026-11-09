package postgres

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// testDB opens the database named in ../../.env; the test is skipped when the
// file or the server is not available (CI without Postgres).
func testDB(t *testing.T) *sqlx.DB {
	t.Helper()
	_ = godotenv.Load("../../.env")
	if os.Getenv("DB_NAME") == "" {
		t.Skip("no .env with DB_NAME")
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Skipf("database not reachable: %v", err)
	}
	return db
}

const demo2Company = "3058efec-88cc-4e41-b822-b36977d2df15" // "Demo 2", never used in production flows

// The SQL of SetBlocked/UpsertSetting is only exercised against a real
// database: parameter type inference ($3 in a CASE) is what broke before.
func TestCompanyRepoSQLAgainstDatabase(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	repo := NewCompanyRepo(db)

	if err := repo.SetBlocked(demo2Company, true, 8); err != nil {
		t.Fatalf("block: %v", err)
	}
	show, err := repo.GetShow(demo2Company)
	if err != nil || !show.IsBlocked {
		t.Fatalf("after block: err=%v blocked=%v", err, show.IsBlocked)
	}
	var blockedBy *int64
	if err := db.Get(&blockedBy, `select blocked_by from companies where id = $1`, demo2Company); err != nil || blockedBy == nil || *blockedBy != 8 {
		t.Fatalf("blocked_by: err=%v value=%v", err, blockedBy)
	}
	if err := repo.SetBlocked(demo2Company, false, 8); err != nil {
		t.Fatalf("unblock: %v", err)
	}
	show, _ = repo.GetShow(demo2Company)
	if show.IsBlocked {
		t.Fatal("still blocked after unblock")
	}

	if err := repo.UpsertSetting(demo2Company, "kpi_enabled", json.RawMessage("true"), 8); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	defer db.Exec(`delete from company_settings where company_id = $1 and key = 'kpi_enabled'`, demo2Company)
	show, _ = repo.GetShow(demo2Company)
	if string(show.Settings["kpi_enabled"]) != "true" {
		t.Fatalf("setting not stored: %s", show.Settings["kpi_enabled"])
	}
}

func TestCompanyUsersCountExcludesDeleted(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() }) // insertTestUser cleanup'idan keyin yopiladi
	repo := NewCompanyRepo(db)

	before, err := repo.GetShow(demo2Company)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	userID := insertTestUser(t, db, "Test count")
	if _, err := db.Exec(`insert into user_companies (user_id, company_id, permission_ids) values ($1, $2, '{}')`, userID, demo2Company); err != nil {
		t.Fatalf("insert user_companies: %v", err)
	}
	if _, err := db.Exec(`update users set deleted_at = now() where id = $1`, userID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	after, err := repo.GetShow(demo2Company)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if after.UsersCount != before.UsersCount {
		t.Fatalf("users_count = %d, want %d (deleted user counted)", after.UsersCount, before.UsersCount)
	}
}
