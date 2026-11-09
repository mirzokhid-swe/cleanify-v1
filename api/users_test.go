package api

import (
	"database/sql"
	"net/http"
	"strings"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

type fakeUserRepo struct {
	repo.UserI
	createErr  error
	created    *models.CreateUserModel
	edited     *models.UpdateUserRequest
	target     *models.User // GetById result; nil → sql.ErrNoRows
	deletedID  int64
	deletedBy  int64
	deleteCall bool
}

func (f *fakeUserRepo) Create(e models.CreateUserModel) (int64, error) {
	if f.createErr != nil {
		return 0, f.createErr
	}
	f.created = &e
	return 9, nil
}

func (f *fakeUserRepo) GetById(id int64) (models.User, error) {
	if f.target == nil {
		return models.User{}, sql.ErrNoRows
	}
	return *f.target, nil
}

func (f *fakeUserRepo) Edit(e models.UpdateUserRequest) (int64, error) {
	f.edited = &e
	return 1, nil
}

func (f *fakeUserRepo) Delete(id int64, deletedBy int64) (int64, error) {
	f.deleteCall = true
	f.deletedID, f.deletedBy = id, deletedBy
	return 1, nil
}

func newUsersRouter(user gin.HandlerFunc, users *fakeUserRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handlers.Handler{Stg: usersStorage{users: users, roles: &fakeRoleRepo{}}}
	r := gin.New()
	usersRoutes(r.Group("/users"), &h, user)
	return r
}

func targetIn(companyID string) *models.User {
	return &models.User{ID: 9, UserPermissionByCompany: []models.UserPermissionByCompany{{CompanyID: companyID}}}
}

func TestUsersCreateRequiresUserCreate(t *testing.T) {
	users := &fakeUserRepo{}
	r := newUsersRouter(userWith("edit_user"), users)
	w := do(r, "POST", "/users", `{"company_id":"`+companyA+`","fullname":"Ali","phone":"991234567"}`)
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" || users.created != nil {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUsersCreateRejectsForeignRole(t *testing.T) {
	users := &fakeUserRepo{createErr: models.ErrRoleForeign}
	r := newUsersRouter(userWith("user_create"), users)
	w := do(r, "POST", "/users", `{"company_id":"`+companyA+`","fullname":"Ali","phone":"991234567","permissions":[{"company_id":"`+companyA+`","role_id":3}]}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUsersEditDeletedIs404(t *testing.T) {
	users := &fakeUserRepo{}
	r := newUsersRouter(userWith("edit_user"), users)
	w := do(r, "POST", "/users/edit", `{"company_id":"`+companyA+`","id":"9","fullname":"Ali","permissions":[{"company_id":"`+companyA+`"}]}`)
	if w.Code != http.StatusNotFound || users.edited != nil {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUsersEditCompanyMismatchIs400(t *testing.T) {
	// caller may edit in both A and B, but the active company A is missing from permissions[]
	users := &fakeUserRepo{target: targetIn(companyB)}
	r := newUsersRouter(func(c *gin.Context) {
		c.Set("user", models.User{ID: 7, UserPermissionByCompany: []models.UserPermissionByCompany{
			{CompanyID: companyA, Can: "edit_user"},
			{CompanyID: companyB, Can: "edit_user"},
		}})
	}, users)
	w := do(r, "POST", "/users/edit", `{"company_id":"`+companyA+`","id":"9","fullname":"Ali","permissions":[{"company_id":"`+companyB+`"}]}`)
	if w.Code != http.StatusBadRequest || users.edited != nil {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUsersEditOK(t *testing.T) {
	users := &fakeUserRepo{target: targetIn(companyA)}
	r := newUsersRouter(userWith("edit_user"), users)
	w := do(r, "POST", "/users/edit", `{"company_id":"`+companyA+`","id":"9","fullname":"Ali","permissions":[{"company_id":"`+companyA+`"}]}`)
	if w.Code != http.StatusOK || users.edited == nil {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUsersDeleteRequiresDeleteUser(t *testing.T) {
	users := &fakeUserRepo{target: targetIn(companyA)}
	r := newUsersRouter(userWith("edit_user"), users)
	w := do(r, "DELETE", "/users/9?company_id="+companyA, "")
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" || users.deleteCall {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUsersDeleteSelfIs400(t *testing.T) {
	users := &fakeUserRepo{target: targetIn(companyA)}
	r := newUsersRouter(userWith("delete_user"), users)
	w := do(r, "DELETE", "/users/7?company_id="+companyA, "")
	if w.Code != http.StatusBadRequest || users.deleteCall {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUsersDeleteOtherCompanyIs404(t *testing.T) {
	users := &fakeUserRepo{target: targetIn(companyB)}
	r := newUsersRouter(userWith("delete_user"), users)
	w := do(r, "DELETE", "/users/9?company_id="+companyA, "")
	if w.Code != http.StatusNotFound || users.deleteCall {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUsersDeleteOK(t *testing.T) {
	users := &fakeUserRepo{target: targetIn(companyA)}
	r := newUsersRouter(userWith("delete_user"), users)
	w := do(r, "DELETE", "/users/9?company_id="+companyA, "")
	if w.Code != http.StatusOK || !users.deleteCall || users.deletedID != 9 || users.deletedBy != 7 {
		t.Fatalf("code=%d body=%s deleted=%d by=%d", w.Code, w.Body.String(), users.deletedID, users.deletedBy)
	}
}

func TestUsersCreateRejectsForeignCompanyInPermissions(t *testing.T) {
	users := &fakeUserRepo{}
	r := newUsersRouter(userWith("user_create"), users)
	w := do(r, "POST", "/users", `{"company_id":"`+companyA+`","fullname":"Ali","phone":"991234567","permissions":[{"company_id":"`+companyA+`"},{"company_id":"`+companyB+`"}]}`)
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" || users.created != nil {
		t.Fatalf("code=%d body=%s created=%v", w.Code, w.Body.String(), users.created)
	}
}

func TestUsersEditRejectsForeignCompanyInPermissions(t *testing.T) {
	users := &fakeUserRepo{target: targetIn(companyA)}
	r := newUsersRouter(userWith("edit_user"), users)
	w := do(r, "POST", "/users/edit", `{"company_id":"`+companyA+`","id":"9","fullname":"Ali","permissions":[{"company_id":"`+companyA+`"},{"company_id":"`+companyB+`"}]}`)
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" || users.edited != nil {
		t.Fatalf("code=%d body=%s edited=%v", w.Code, w.Body.String(), users.edited)
	}
}

func TestUsersEditTargetOutsideCompanyIs404(t *testing.T) {
	users := &fakeUserRepo{target: targetIn(companyB)}
	r := newUsersRouter(userWith("edit_user"), users)
	w := do(r, "POST", "/users/edit", `{"company_id":"`+companyA+`","id":"9","fullname":"Ali","permissions":[{"company_id":"`+companyA+`"}]}`)
	if w.Code != http.StatusNotFound || users.edited != nil {
		t.Fatalf("code=%d body=%s edited=%v", w.Code, w.Body.String(), users.edited)
	}
}

func TestUsersEditAddsSecondCompany(t *testing.T) {
	users := &fakeUserRepo{target: targetIn(companyA)}
	r := newUsersRouter(func(c *gin.Context) {
		c.Set("user", models.User{ID: 7, UserPermissionByCompany: []models.UserPermissionByCompany{
			{CompanyID: companyA, Can: "edit_user"},
			{CompanyID: companyB, Can: "edit_user"},
		}})
	}, users)
	w := do(r, "POST", "/users/edit", `{"company_id":"`+companyA+`","id":"9","fullname":"Ali","permissions":[{"company_id":"`+companyA+`"},{"company_id":"`+companyB+`"}]}`)
	if w.Code != http.StatusOK || users.edited == nil || len(users.edited.Permissions) != 2 {
		t.Fatalf("code=%d body=%s edited=%v", w.Code, w.Body.String(), users.edited)
	}
}

func TestUsersCreateCompanyMismatchIs400(t *testing.T) {
	users := &fakeUserRepo{}
	r := newUsersRouter(userWith("user_create"), users)
	w := do(r, "POST", "/users", `{"company_id":"`+companyA+`","fullname":"Ali","phone":"991234567"}`)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "company_id mismatch") || users.created != nil {
		t.Fatalf("code=%d body=%s created=%v", w.Code, w.Body.String(), users.created)
	}
}

func TestUsersGetByIdDeletedIs404(t *testing.T) {
	users := &fakeUserRepo{}
	r := newUsersRouter(userWith("get_users_list"), users)
	w := do(r, "GET", "/users/9", "")
	if w.Code != http.StatusNotFound || responseStatus(w.Body.Bytes()) != "NOT_FOUND" {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
