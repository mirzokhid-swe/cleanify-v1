package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/storage"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

type fakePermissionRepo struct{ repo.PermissionI }

func (fakePermissionRepo) GetList(scope string) ([]models.Permission, error) {
	return []models.Permission{
		{ID: "1", Slug: "get_orders_list", Name: "Ro'yxatni ko'rish", Group: "Buyurtmalar"},
		{ID: "14", Slug: "change_status_to_6", Name: "x", Group: "Buyurtmalar"},
	}, nil
}

type fakeStatusRepo struct{ repo.OrderStatusI }

func (fakeStatusRepo) GetList(companyID, sortBy, sortOrder string) ([]models.OrderStatus, error) {
	return []models.OrderStatus{{Number: 6, Name: "Topshirildi"}}, nil
}

type fakePermissionStorage struct{ storage.StorageI }

func (fakePermissionStorage) Permission() repo.PermissionI   { return fakePermissionRepo{} }
func (fakePermissionStorage) OrderStatus() repo.OrderStatusI { return fakeStatusRepo{} }

func TestPermissionsGroupedWithCompanyStatuses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.Handler{Stg: fakePermissionStorage{}}
	r := gin.New()
	r.GET("/permissions", h.GetPermissionList)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/permissions?company_id="+companyA, nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var resp struct {
		Data []models.PermissionGroup `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 2 || resp.Data[0].Group != "Buyurtmalar" || resp.Data[1].Group != "Holatni o'zgartirish" {
		t.Fatal(w.Body.String())
	}
	if resp.Data[1].Items[0].Name != "Topshirildi" {
		t.Fatal(resp.Data[1])
	}
}
