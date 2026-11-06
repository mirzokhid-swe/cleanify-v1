package middleware

import (
	"net/http/httptest"
	"testing"

	"bw-erp/models"

	"github.com/gin-gonic/gin"
)

func TestSuperAdminMiddleware(t *testing.T) {
	for _, tc := range []struct {
		name    string
		user    *models.User
		code    int
		aborted bool
	}{
		{"regular user", &models.User{IsSuperAdmin: false}, 403, true},
		{"no user in context", nil, 403, true},
		{"super admin", &models.User{IsSuperAdmin: true}, 200, false},
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/company", nil)
		if tc.user != nil {
			c.Set("user", *tc.user)
		}
		SuperAdminMiddleware()(c)
		if c.IsAborted() != tc.aborted || w.Code != tc.code {
			t.Fatalf("%s: aborted=%v code=%d", tc.name, c.IsAborted(), w.Code)
		}
	}
}

func TestAllCompaniesBlocked(t *testing.T) {
	u := models.User{UserPermissionByCompany: []models.UserPermissionByCompany{{IsBlocked: true}, {IsBlocked: false}}}
	if allCompaniesBlocked(u) {
		t.Fatal("one active company")
	}
	u.UserPermissionByCompany[1].IsBlocked = true
	if !allCompaniesBlocked(u) {
		t.Fatal("all blocked")
	}
	if allCompaniesBlocked(models.User{}) {
		t.Fatal("no companies")
	}
}
