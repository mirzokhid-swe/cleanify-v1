package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"bw-erp/models"

	"github.com/google/uuid"
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

// Create writes the company and every default in one transaction.
func TestCompanyCreateAgainstDatabase(t *testing.T) {
	db := testDB(t)
	// Cleanups run last-in-first-out: the database closes after the rows are removed.
	t.Cleanup(func() { db.Close() })
	repo := NewCompanyRepo(db)

	id := uuid.New().String()
	name := "test-create-" + id
	cleanup := func(companyID string) {
		db.Exec(`delete from user_companies where company_id = $1`, companyID)
		db.Exec(`delete from roles where company_id = $1::uuid`, companyID)
		db.Exec(`delete from payment_purposes where company_id = $1::uuid`, companyID)
		db.Exec(`delete from order_item_types where company_id = $1::uuid`, companyID)
		db.Exec(`delete from order_statuses where company_id = $1::uuid`, companyID)
		db.Exec(`delete from companies where id = $1::uuid`, companyID)
	}
	t.Cleanup(func() { cleanup(id) })

	if err := repo.Create(id, name, 8); err != nil {
		t.Fatalf("create: %v", err)
	}

	count := func(query string) int {
		var n int
		if err := db.Get(&n, query, id); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
		return n
	}
	if n := count(`select count(*) from order_statuses where company_id = $1::uuid and coalesce(color, '') <> ''`); n != 7 {
		t.Fatalf("colored statuses: %d", n)
	}
	if n := count(`select count(*) from order_statuses where company_id = $1::uuid and number = 8 and slug = 'change_status_to_8' and "order" = 7`); n != 1 {
		t.Fatalf("Omborda (8): %d", n)
	}
	if n := count(`select count(*) from order_item_types where company_id = $1::uuid`); n != 8 {
		t.Fatalf("types: %d", n)
	}
	if n := count(`select count(*) from order_item_types where company_id = $1::uuid and
		((name = 'Parda' and unit = 'weight' and price = 30000 and not is_countable) or
		 (name = 'Ko''rpa' and unit = 'length' and price = 25000 and not is_countable) or
		 (name = 'Odeyalo 50' and unit = 'piece' and price = 50000 and is_countable))`); n != 3 {
		t.Fatalf("Parda/Ko'rpa/Odeyalo 50: %d", n)
	}
	if n := count(`select count(*) from roles where company_id = $1::uuid`); n != 4 {
		t.Fatalf("roles: %d", n)
	}
	if n := count(`select count(*) from roles where company_id = $1::uuid and name = 'Admin'
		and cardinality(permission_ids) = (select count(*) from permissions where scope = 'company')`); n != 1 {
		t.Fatalf("Admin role with every company permission: %d", n)
	}
	if n := count(`select cardinality(permission_ids) from roles where company_id = $1::uuid and name = 'Kuryer'`); n != 8 {
		t.Fatalf("Kuryer permissions: %d", n)
	}
	if n := count(`select count(*) from payment_purposes where company_id = $1::uuid`); n != 3 {
		t.Fatalf("payment purposes: %d", n)
	}
	if n := count(`select count(*) from user_companies uc join roles r on r.id = uc.role_id
		where uc.company_id = $1 and uc.user_id = 8 and r.name = 'Admin'`); n != 1 {
		t.Fatalf("creator as Admin: %d", n)
	}

	// The same name again: an error and nothing left behind.
	second := uuid.New().String()
	t.Cleanup(func() { cleanup(second) })
	if err := repo.Create(second, name, 8); !errors.Is(err, models.ErrCompanyNameTaken) {
		t.Fatalf("duplicate name: want ErrCompanyNameTaken, got %v", err)
	}
	var left int
	db.Get(&left, `select (select count(*) from companies where id = $1::uuid) + (select count(*) from order_statuses where company_id = $1::uuid)`, second)
	if left != 0 {
		t.Fatalf("duplicate left %d rows", left)
	}
}
