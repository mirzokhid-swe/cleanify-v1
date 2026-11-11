package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/storage"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

type fakeRoleRepo struct {
	repo.RoleI
	created     *models.RoleInput
	createErr   error
	listCompany string
	deleteErr   error
	deleteRows  int64
	usersCount  int
}

func (f *fakeRoleRepo) GetList(companyID string) ([]models.Role, error) {
	f.listCompany = companyID
	return []models.Role{{ID: 1, CompanyID: companyID, Name: "Kassir"}}, nil
}

func (f *fakeRoleRepo) GetByID(id int64, companyID string) (models.Role, error) {
	if companyID != companyA {
		return models.Role{}, sql.ErrNoRows
	}
	return models.Role{ID: id, CompanyID: companyA, Name: "Kassir", UsersCount: f.usersCount}, nil
}

func (f *fakeRoleRepo) Create(in models.RoleInput) (int64, error) {
	if f.createErr != nil {
		return 0, f.createErr
	}
	f.created = &in
	return 5, nil
}

func (f *fakeRoleRepo) Delete(id int64, companyID string) (int64, error) {
	return f.deleteRows, f.deleteErr
}

type usersStorage struct {
	storage.StorageI
	users *fakeUserRepo
	roles *fakeRoleRepo
}

func (s usersStorage) User() repo.UserI { return s.users }
func (s usersStorage) Role() repo.RoleI { return s.roles }

// userWith: user 7, works in company A with the given slugs.
func userWith(can string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", models.User{ID: 7, UserPermissionByCompany: []models.UserPermissionByCompany{{
			CompanyID: companyA, Can: can,
		}}})
	}
}

func newRolesRouter(user gin.HandlerFunc, roles *fakeRoleRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handlers.Handler{Stg: usersStorage{users: &fakeUserRepo{}, roles: roles}}
	r := gin.New()
	rolesRoutes(r.Group("/roles"), &h, user)
	return r
}

func TestRolesWriteRequireManageRoles(t *testing.T) {
	roles := &fakeRoleRepo{}
	r := newRolesRouter(userWith("edit_user"), roles)
	w := do(r, "POST", "/roles", `{"company_id":"`+companyA+`","name":"Kassir"}`)
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" || roles.created != nil {
		t.Fatalf("code=%d body=%s created=%v", w.Code, w.Body.String(), roles.created)
	}
}

func TestRolesListAllowedWithEditUser(t *testing.T) {
	roles := &fakeRoleRepo{}
	r := newRolesRouter(userWith("edit_user"), roles)
	w := do(r, "GET", "/roles?company_id="+companyA, "")
	if w.Code != http.StatusOK || roles.listCompany != companyA {
		t.Fatalf("code=%d body=%s listCompany=%q", w.Code, w.Body.String(), roles.listCompany)
	}
	var resp struct {
		Data []models.Role `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) != 1 || resp.Data[0].Name != "Kassir" {
		t.Fatalf("data=%+v", resp.Data)
	}
}

func TestRolesListDeniedWithoutAnySlug(t *testing.T) {
	roles := &fakeRoleRepo{}
	r := newRolesRouter(userWith("get_orders_list"), roles)
	w := do(r, "GET", "/roles?company_id="+companyA, "")
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" || roles.listCompany != "" {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	w = do(r, "GET", "/roles/1?company_id="+companyA, "")
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" {
		t.Fatalf("show: code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRolesShowOtherCompanyIs404(t *testing.T) {
	r := newRolesRouter(func(c *gin.Context) {
		c.Set("user", models.User{ID: 7, UserPermissionByCompany: []models.UserPermissionByCompany{
			{CompanyID: companyB, Can: "manage_roles"},
		}})
	}, &fakeRoleRepo{})
	w := do(r, "GET", "/roles/1?company_id="+companyB, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRoleDeleteInUseIs409(t *testing.T) {
	roles := &fakeRoleRepo{deleteErr: models.ErrRoleInUse, usersCount: 3}
	r := newRolesRouter(userWith("manage_roles"), roles)
	w := do(r, "DELETE", "/roles/1?company_id="+companyA, "")
	if w.Code != http.StatusConflict || responseStatus(w.Body.Bytes()) != "CONFLICT" {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			UsersCount int `json:"users_count"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.UsersCount != 3 {
		t.Fatalf("users_count=%d body=%s", resp.Data.UsersCount, w.Body.String())
	}
}

func TestRoleDeleteUnknownIs404(t *testing.T) {
	r := newRolesRouter(userWith("manage_roles"), &fakeRoleRepo{deleteRows: 0})
	w := do(r, "DELETE", "/roles/1?company_id="+companyA, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRoleCreateDuplicateIs409(t *testing.T) {
	roles := &fakeRoleRepo{createErr: models.ErrRoleNameTaken}
	r := newRolesRouter(userWith("manage_roles"), roles)
	w := do(r, "POST", "/roles", `{"company_id":"`+companyA+`","name":"Kassir"}`)
	if w.Code != http.StatusConflict || responseStatus(w.Body.Bytes()) != "CONFLICT" {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRoleCreateOK(t *testing.T) {
	roles := &fakeRoleRepo{}
	r := newRolesRouter(userWith("manage_roles"), roles)
	w := do(r, "POST", "/roles", `{"company_id":"`+companyA+`","name":"Kassir","permission_ids":[1,2]}`)
	if w.Code != http.StatusCreated || roles.created == nil || roles.created.CompanyID != companyA {
		t.Fatalf("code=%d body=%s created=%+v", w.Code, w.Body.String(), roles.created)
	}
}
