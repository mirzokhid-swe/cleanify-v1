package api

import (
	"net/http"
	"strings"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/pkg/utils"

	"github.com/gin-gonic/gin"
)

func newRefreshRouter(users *fakeUserRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handlers.Handler{Stg: usersStorage{users: users, roles: &fakeRoleRepo{}}}
	r := gin.New()
	r.POST("/auth/refresh-token", h.RefreshToken)
	return r
}

func TestRefreshTokenDeletedUser401(t *testing.T) {
	t.Setenv("API_SECRET", "test-secret")
	_, refresh, err := utils.GenerateToken("9", "998901234567")
	if err != nil {
		t.Fatal(err)
	}
	r := newRefreshRouter(&fakeUserRepo{}) // GetById → sql.ErrNoRows (o'chirilgan)
	w := do(r, "POST", "/auth/refresh-token", `{"refresh_token":"`+refresh+`"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRefreshTokenPhoneMismatch401(t *testing.T) {
	t.Setenv("API_SECRET", "test-secret")
	_, refresh, err := utils.GenerateToken("9", "998901234567")
	if err != nil {
		t.Fatal(err)
	}
	r := newRefreshRouter(&fakeUserRepo{target: &models.User{ID: 9, Phone: "998900000000"}})
	w := do(r, "POST", "/auth/refresh-token", `{"refresh_token":"`+refresh+`"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRefreshTokenBySub(t *testing.T) {
	t.Setenv("API_SECRET", "test-secret")
	_, refresh, err := utils.GenerateToken("9", "998901234567")
	if err != nil {
		t.Fatal(err)
	}
	r := newRefreshRouter(&fakeUserRepo{target: &models.User{ID: 9, Phone: "998901234567"}})
	w := do(r, "POST", "/auth/refresh-token", `{"refresh_token":"`+refresh+`"}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "access_token") {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
