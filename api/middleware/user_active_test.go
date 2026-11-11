package middleware

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"bw-erp/models"
	"bw-erp/pkg/utils"
	"bw-erp/storage"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

type fakeUserRepo struct {
	repo.UserI
}

func (fakeUserRepo) GetById(id int64) (models.User, error) {
	return models.User{}, sql.ErrNoRows
}

type fakeStorage struct {
	storage.StorageI
}

func (fakeStorage) User() repo.UserI {
	return fakeUserRepo{}
}

func TestUserActiveMiddlewareDeletedUser401(t *testing.T) {
	t.Setenv("API_SECRET", "test-secret")
	token, _, err := utils.GenerateToken("42", "998901234567")
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", AuthMiddleware(), UserActiveMiddleware(fakeStorage{}), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
