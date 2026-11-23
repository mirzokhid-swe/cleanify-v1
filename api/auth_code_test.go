package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

const authCodeUserID = int64(77)

type fakeAuthCodeRepo struct {
	repo.AuthCodeI
	activeCode string
	userID     int64
}

func (f *fakeAuthCodeRepo) Use(code string) (int64, error) {
	if code == f.activeCode {
		f.activeCode = ""
		return f.userID, nil
	}
	return 0, sql.ErrNoRows
}

type fakeAuthUserRepo struct {
	repo.UserI
	deleted bool
}

func (f fakeAuthUserRepo) GetById(id int64) (models.User, error) {
	if f.deleted || id != authCodeUserID {
		return models.User{}, sql.ErrNoRows
	}
	return models.User{ID: id, Phone: "+998901234567"}, nil
}

type fakeAuthStorage struct {
	fakeStorage
	codes *fakeAuthCodeRepo
	users fakeAuthUserRepo
}

func (f fakeAuthStorage) AuthCode() repo.AuthCodeI { return f.codes }
func (f fakeAuthStorage) User() repo.UserI         { return f.users }

func newAuthCodeRouter(t *testing.T, deletedUser bool) *gin.Engine {
	t.Helper()
	t.Setenv("API_SECRET", "test-secret")
	handlers.ResetAuthCodeLimiter()
	gin.SetMode(gin.TestMode)
	h := handlers.Handler{Stg: fakeAuthStorage{
		codes: &fakeAuthCodeRepo{activeCode: "483920", userID: authCodeUserID},
		users: fakeAuthUserRepo{deleted: deletedUser},
	}}
	r := gin.New()
	r.POST("/auth/verify-code", h.VerifyAuthCode)
	return r
}

func TestVerifyAuthCodeOK(t *testing.T) {
	r := newAuthCodeRouter(t, false)
	w := do(r, "POST", "/auth/verify-code", `{"code":"483920"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data models.AuthorizationResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Data.AccessToken == "" {
		t.Fatalf("empty access_token: %s", w.Body.String())
	}
}

func TestVerifyAuthCodeWrong(t *testing.T) {
	r := newAuthCodeRouter(t, false)
	w := do(r, "POST", "/auth/verify-code", `{"code":"000000"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyAuthCodeDeletedUser(t *testing.T) {
	r := newAuthCodeRouter(t, true)
	w := do(r, "POST", "/auth/verify-code", `{"code":"483920"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyAuthCodeRateLimit(t *testing.T) {
	r := newAuthCodeRouter(t, false)
	for i := 0; i < 10; i++ {
		if w := do(r, "POST", "/auth/verify-code", `{"code":"111111"}`); w.Code != http.StatusBadRequest {
			t.Fatalf("attempt %d: code=%d", i+1, w.Code)
		}
	}
	if w := do(r, "POST", "/auth/verify-code", `{"code":"111111"}`); w.Code != http.StatusTooManyRequests {
		t.Fatalf("11th attempt: code=%d body=%s", w.Code, w.Body.String())
	}
}

// Limiter haqiqiy ulanish manziliga tayanishi kerak: X-Forwarded-For ni
// hujumchi o'zi yozadi, u bilan limit aylanib o'tilmasligi shart.
func TestVerifyAuthCodeRateLimitIgnoresForwardedFor(t *testing.T) {
	r := newAuthCodeRouter(t, false)
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/verify-code", strings.NewReader(`{"code":"111111"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.0.0.%d", i+1))
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("attempt %d: code=%d", i+1, w.Code)
		}
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/verify-code", strings.NewReader(`{"code":"111111"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "10.0.0.99")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("11th attempt with fresh XFF: code=%d, limiter spoofed", w.Code)
	}
}

func TestVerifyAuthCodeBadBody(t *testing.T) {
	r := newAuthCodeRouter(t, false)
	if w := do(r, "POST", "/auth/verify-code", `{"code":"48392"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
