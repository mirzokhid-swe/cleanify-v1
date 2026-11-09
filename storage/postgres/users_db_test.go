package postgres

import (
	"bw-erp/models"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// insertTestUser: test xodimi; o'chirish cleanup'i user_companies bilan birga.
// Rollar xodimdan oldin yaratiladi, shunda ularning cleanup'i keyin ishlaydi.
func insertTestUser(t *testing.T, db *sqlx.DB, fullname string) int64 {
	t.Helper()
	var userID int64
	phone := fmt.Sprintf("+99800%07d", rand.Intn(10000000))
	err := db.Get(&userID, `insert into users (phone, fullname, is_active) values ($1, $2, true) returning id`, phone, fullname)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`delete from user_companies where user_id = $1`, userID)
		db.Exec(`delete from users where id = $1`, userID)
	})
	return userID
}

func createTestRole(t *testing.T, db *sqlx.DB, companyID string, name string, permissionIDs []int) int64 {
	t.Helper()
	id, err := NewRoleRepo(db).Create(models.RoleInput{CompanyID: companyID, Name: name, PermissionIDs: permissionIDs})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	t.Cleanup(func() { db.Exec(`delete from roles where id = $1`, id) })
	return id
}

func TestUserCanFromRole(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewUserRepo(db)

	roleID := createTestRole(t, db, demo2Company, "Test rol can", []int{8})
	userID := insertTestUser(t, db, "Test")
	if _, err := db.Exec(`insert into user_companies (user_id, company_id, permission_ids) values ($1, $2, '{37}')`, userID, demo2Company); err != nil {
		t.Fatalf("insert user_companies: %v", err)
	}

	_, err := repo.Edit(models.UpdateUserRequest{
		ID:          strconv.FormatInt(userID, 10),
		Permissions: []models.UserCompanyInput{{CompanyID: demo2Company, RoleID: &roleID, PermissionIDs: []int{37}}},
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	user, err := repo.GetById(userID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(user.UserPermissionByCompany) != 1 {
		t.Fatalf("companies: %+v", user.UserPermissionByCompany)
	}
	p := user.UserPermissionByCompany[0]
	if p.Can != "get_orders_list" {
		t.Fatalf("can = %q, want get_orders_list", p.Can)
	}
	if p.RoleID == nil || *p.RoleID != roleID || p.RoleName == nil || *p.RoleName != "Test rol can" {
		t.Fatalf("role: id=%v name=%v", p.RoleID, p.RoleName)
	}
	if !reflect.DeepEqual(p.PermissionIDs, []int{37}) {
		t.Fatalf("permission_ids = %v, want [37]", p.PermissionIDs)
	}
}

func TestUserEditWithRoleKeepsPermissionIDs(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewUserRepo(db)

	roleID := createTestRole(t, db, demo2Company, "Test rol keep", []int{8})
	userID := insertTestUser(t, db, "Test")
	if _, err := db.Exec(`insert into user_companies (user_id, company_id, permission_ids) values ($1, $2, '{37}')`, userID, demo2Company); err != nil {
		t.Fatalf("insert user_companies: %v", err)
	}

	_, err := repo.Edit(models.UpdateUserRequest{
		ID:          strconv.FormatInt(userID, 10),
		Permissions: []models.UserCompanyInput{{CompanyID: demo2Company, RoleID: &roleID, PermissionIDs: []int{}}},
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	var ids pq.Int64Array
	var gotRole *int64
	err = db.QueryRow(`select permission_ids, role_id from user_companies where user_id = $1 and company_id = $2`, userID, demo2Company).Scan(&ids, &gotRole)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if !reflect.DeepEqual(ids, pq.Int64Array{37}) {
		t.Fatalf("permission_ids = %v, want {37}", ids)
	}
	if gotRole == nil || *gotRole != roleID {
		t.Fatalf("role_id = %v, want %d", gotRole, roleID)
	}
}

func TestUserEditNullRoleUsesOwnPermissions(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewUserRepo(db)

	roleID := createTestRole(t, db, demo2Company, "Test rol null", []int{8})
	userID := insertTestUser(t, db, "Test")
	if _, err := db.Exec(`insert into user_companies (user_id, company_id, permission_ids, role_id) values ($1, $2, '{}', $3)`, userID, demo2Company, roleID); err != nil {
		t.Fatalf("insert user_companies: %v", err)
	}

	_, err := repo.Edit(models.UpdateUserRequest{
		ID:          strconv.FormatInt(userID, 10),
		Permissions: []models.UserCompanyInput{{CompanyID: demo2Company, RoleID: nil, PermissionIDs: []int{37}}},
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	var gotRole *int64
	if err := db.Get(&gotRole, `select role_id from user_companies where user_id = $1 and company_id = $2`, userID, demo2Company); err != nil {
		t.Fatalf("select: %v", err)
	}
	if gotRole != nil {
		t.Fatalf("role_id = %d, want NULL", *gotRole)
	}

	user, err := repo.GetById(userID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(user.UserPermissionByCompany) != 1 {
		t.Fatalf("companies: %+v", user.UserPermissionByCompany)
	}
	p := user.UserPermissionByCompany[0]
	if p.Can != "delete_order" || p.RoleID != nil || p.RoleName != nil {
		t.Fatalf("can=%q role=%v name=%v", p.Can, p.RoleID, p.RoleName)
	}
}

func TestUserEditRejectsForeignRole(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewUserRepo(db)

	var otherCompany string
	if err := db.Get(&otherCompany, `select id from companies where id <> $1 limit 1`, demo2Company); err != nil {
		t.Skipf("boshqa korxona yo'q: %v", err)
	}
	roleID := createTestRole(t, db, otherCompany, "Test rol foreign", []int{8})
	userID := insertTestUser(t, db, "Test")

	_, err := repo.Edit(models.UpdateUserRequest{
		ID:          strconv.FormatInt(userID, 10),
		Permissions: []models.UserCompanyInput{{CompanyID: demo2Company, RoleID: &roleID}},
	})
	if !errors.Is(err, models.ErrRoleForeign) {
		t.Fatalf("want ErrRoleForeign, got %v", err)
	}
	var n int
	db.Get(&n, `select count(*) from user_companies where user_id = $1`, userID)
	if n != 0 {
		t.Fatalf("user_companies rows written: %d", n)
	}
}

func TestUserDeleteHidesFromListAndGetById(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewUserRepo(db)

	roleID := createTestRole(t, db, demo2Company, "Test rol delete", []int{8})
	userID := insertTestUser(t, db, "Test")
	if _, err := db.Exec(`insert into user_companies (user_id, company_id, permission_ids, role_id, is_courier) values ($1, $2, '{37}', $3, true)`, userID, demo2Company, roleID); err != nil {
		t.Fatalf("insert user_companies: %v", err)
	}

	inList := func() bool {
		list, err := repo.GetList(demo2Company)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for _, u := range list {
			if u.ID == userID {
				return true
			}
		}
		return false
	}
	inCouriers := func() bool {
		couriers, err := repo.GetCouriesList(demo2Company)
		if err != nil {
			t.Fatalf("couriers: %v", err)
		}
		for _, c := range couriers {
			if c.ID == strconv.FormatInt(userID, 10) {
				return true
			}
		}
		return false
	}
	if !inList() || !inCouriers() {
		t.Fatalf("before delete: inList=%v inCouriers=%v", inList(), inCouriers())
	}

	n, err := repo.Delete(userID, 8)
	if err != nil || n != 1 {
		t.Fatalf("delete: n=%d err=%v", n, err)
	}

	if inList() {
		t.Fatal("deleted user still in GetList")
	}
	if inCouriers() {
		t.Fatal("deleted user still in GetCouriesList")
	}
	if _, err := repo.GetById(userID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetById: want sql.ErrNoRows, got %v", err)
	}

	var deletedBy *int64
	var isActive bool
	var chatID *int64
	err = db.QueryRow(`select deleted_by, is_active, chat_id from users where id = $1 and deleted_at is not null`, userID).Scan(&deletedBy, &isActive, &chatID)
	if err != nil || deletedBy == nil || *deletedBy != 8 || isActive || chatID != nil {
		t.Fatalf("users row: err=%v deleted_by=%v is_active=%v chat_id=%v", err, deletedBy, isActive, chatID)
	}
	var ids pq.Int64Array
	var gotRole *int64
	err = db.QueryRow(`select permission_ids, role_id from user_companies where user_id = $1`, userID).Scan(&ids, &gotRole)
	if err != nil || gotRole != nil || !reflect.DeepEqual(ids, pq.Int64Array{37}) {
		t.Fatalf("user_companies: err=%v role_id=%v permission_ids=%v", err, gotRole, ids)
	}

	n, err = repo.Delete(userID, 8)
	if err != nil || n != 0 {
		t.Fatalf("second delete: n=%d err=%v", n, err)
	}
}

func TestUserListFromUserCompanies(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })
	repo := NewUserRepo(db)

	var otherCompany string
	if err := db.Get(&otherCompany, `select id from companies where id <> $1 limit 1`, demo2Company); err != nil {
		t.Skipf("boshqa korxona yo'q: %v", err)
	}
	roleID := createTestRole(t, db, demo2Company, "Test rol list", []int{8})

	phone := fmt.Sprintf("+99800%07d", rand.Intn(10000000))
	userID, err := repo.Create(models.CreateUserModel{
		Fullname:    "Test list",
		Phone:       phone,
		CompanyID:   otherCompany,
		Permissions: []models.UserCompanyInput{{CompanyID: demo2Company, RoleID: &roleID, PermissionIDs: []int{37}, IsCourier: true}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`delete from user_companies where user_id = $1`, userID)
		db.Exec(`delete from users where id = $1`, userID)
	})

	list, err := repo.GetList(demo2Company)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var got *models.UserListItem
	for i := range list {
		if list[i].ID == userID {
			got = &list[i]
		}
	}
	if got == nil {
		t.Fatalf("user %d not in demo2 list", userID)
	}
	if got.Fullname != "Test list" || got.Phone != phone || !got.IsActive || !got.IsCourier {
		t.Fatalf("listed user: %+v", got)
	}
	if got.RoleID == nil || *got.RoleID != roleID || got.RoleName == nil || *got.RoleName != "Test rol list" {
		t.Fatalf("role: id=%v name=%v", got.RoleID, got.RoleName)
	}
	// rol bilan yaratilganda o'z huquqlari yozilmaydi
	if len(got.PermissionIDs) != 0 {
		t.Fatalf("permission_ids = %v, want empty", got.PermissionIDs)
	}

	otherList, err := repo.GetList(otherCompany)
	if err != nil {
		t.Fatalf("other list: %v", err)
	}
	for _, u := range otherList {
		if u.ID == userID {
			t.Fatal("user listed by users.company_id, want only user_companies")
		}
	}
}

func TestOrderGetByIdCourierNameAfterDelete(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() { db.Close() })

	courierID := insertTestUser(t, db, "Test kuryer")
	var orderID int
	if err := db.Get(&orderID, `insert into orders (company_id, courier_id, status, count) values ($1, $2, 1, 1) returning id`, demo2Company, courierID); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	t.Cleanup(func() { db.Exec(`delete from orders where id = $1`, orderID) })

	if _, err := NewUserRepo(db).Delete(courierID, 8); err != nil {
		t.Fatalf("delete courier: %v", err)
	}

	order, err := NewOrderRepo(db).GetDetailedByPrimaryKey(orderID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if order.CourierName == nil || *order.CourierName != "Test kuryer" {
		t.Fatalf("courier_name = %v, want Test kuryer", order.CourierName)
	}
}
