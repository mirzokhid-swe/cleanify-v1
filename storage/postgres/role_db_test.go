package postgres

import (
	"bw-erp/models"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestRoleRepoCRUD(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewRoleRepo(db)

	id, err := repo.Create(models.RoleInput{CompanyID: demo2Company, Name: "Test rol", PermissionIDs: []int{1, 2}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { db.Exec(`delete from roles where id = $1`, id) })

	list, err := repo.GetList(demo2Company)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	found := false
	for _, r := range list {
		if r.ID == id {
			found = true
			if r.Name != "Test rol" || r.UsersCount != 0 || !reflect.DeepEqual(r.PermissionIDs, []int{1, 2}) {
				t.Fatalf("listed role: %+v", r)
			}
		}
	}
	if !found {
		t.Fatalf("role %d not in list", id)
	}

	n, err := repo.Update(id, models.RoleInput{CompanyID: demo2Company, Name: "Test rol 2", PermissionIDs: []int{1}})
	if err != nil || n != 1 {
		t.Fatalf("update: n=%d err=%v", n, err)
	}
	got, err := repo.GetByID(id, demo2Company)
	if err != nil || got.Name != "Test rol 2" || !reflect.DeepEqual(got.PermissionIDs, []int{1}) {
		t.Fatalf("get: %+v err=%v", got, err)
	}

	_, err = repo.Create(models.RoleInput{CompanyID: demo2Company, Name: "Test rol 2"})
	if !errors.Is(err, models.ErrRoleNameTaken) {
		t.Fatalf("duplicate create: want ErrRoleNameTaken, got %v", err)
	}

	_, err = repo.GetByID(id, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("other company: want sql.ErrNoRows, got %v", err)
	}

	n, err = repo.Delete(id, demo2Company)
	if err != nil || n != 1 {
		t.Fatalf("delete: n=%d err=%v", n, err)
	}
	list, _ = repo.GetList(demo2Company)
	for _, r := range list {
		if r.ID == id {
			t.Fatalf("role %d still in list after delete", id)
		}
	}
}

func TestRoleDeleteInUse(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewRoleRepo(db)

	id, err := repo.Create(models.RoleInput{CompanyID: demo2Company, Name: "Test rol band"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { db.Exec(`delete from roles where id = $1`, id) })

	var userID int64
	phone := fmt.Sprintf("+99800%07d", rand.Intn(10000000))
	err = db.Get(&userID, `insert into users (phone, fullname, is_active) values ($1, 'Test', true) returning id`, phone)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() { db.Exec(`delete from users where id = $1`, userID) })

	_, err = db.Exec(`insert into user_companies (user_id, company_id, permission_ids, role_id) values ($1, $2, '{}', $3)`, userID, demo2Company, id)
	if err != nil {
		t.Fatalf("insert user_companies: %v", err)
	}
	t.Cleanup(func() { db.Exec(`delete from user_companies where user_id = $1`, userID) })

	_, err = repo.Delete(id, demo2Company)
	if !errors.Is(err, models.ErrRoleInUse) {
		t.Fatalf("want ErrRoleInUse, got %v", err)
	}
	got, err := repo.GetByID(id, demo2Company)
	if err != nil || got.UsersCount != 1 {
		t.Fatalf("role after refused delete: %+v err=%v", got, err)
	}
}
