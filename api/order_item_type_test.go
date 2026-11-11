package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/storage"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

const (
	companyA = "11111111-1111-1111-1111-111111111111"
	companyB = "22222222-2222-2222-2222-222222222222"
	typeOfB  = "33333333-3333-3333-3333-333333333333"
)

type fakeTypeRepo struct {
	repo.OrderItemTypeI
	created    *models.OrderItemTypeModel
	updated    *models.EditOrderItemTypeRequest
	gotCompany string
	delCompany string
}

func (f *fakeTypeRepo) Create(id string, e models.OrderItemTypeModel) error {
	f.created = &e
	return nil
}

func (f *fakeTypeRepo) Update(e models.EditOrderItemTypeRequest) (int64, error) {
	f.updated = &e
	return 1, nil
}

func (f *fakeTypeRepo) GetById(id string, companyID string) (models.OrderItemByCompany, error) {
	f.gotCompany = companyID
	if companyID == "" || companyID == companyB {
		return models.OrderItemByCompany{ID: id, Name: "B's type", CopmanyID: companyB}, nil
	}
	return models.OrderItemByCompany{}, sql.ErrNoRows
}

func (f *fakeTypeRepo) Delete(id string, companyID string) (int64, error) {
	f.delCompany = companyID
	return 1, nil
}

type fakeCompanyRepo struct {
	repo.CompanyStorageI
}

func (fakeCompanyRepo) GetById(id string) (models.Company, error) {
	return models.Company{}, nil
}

type fakeStorage struct {
	storage.StorageI
	types *fakeTypeRepo
}

func (f fakeStorage) OrderItemType() repo.OrderItemTypeI { return f.types }
func (f fakeStorage) Company() repo.CompanyStorageI      { return fakeCompanyRepo{} }

// userA is an employee of company A with every type permission there, and nothing in B.
func userA(c *gin.Context) {
	c.Set("user", models.User{UserPermissionByCompany: []models.UserPermissionByCompany{{
		CompanyID: companyA,
		Can:       "get_order_item_type|create_order_item_type|edit_order_item_type|delete_order_item_type",
	}}})
}

func newTypeRouter() (*gin.Engine, *fakeTypeRepo) {
	gin.SetMode(gin.TestMode)
	types := &fakeTypeRepo{}
	h := handlers.Handler{Stg: fakeStorage{types: types}}
	r := gin.New()
	orderItemTypeRoutes(r.Group("/order-item-type"), &h, userA)
	return r, types
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestCreateUsesVerifiedCompanyOnly(t *testing.T) {
	r, types := newTypeRouter()
	w := do(r, "POST", "/order-item-type?company_id="+companyA, `{"company_id":"`+companyB+`","name":"Gilam","price":10,"unit":"area"}`)
	if w.Code != http.StatusForbidden || types.created != nil {
		t.Fatalf("code=%d created=%v", w.Code, types.created)
	}
	w = do(r, "POST", "/order-item-type", `{"company_id":"`+companyA+`","name":"Gilam","price":10,"unit":"area"}`)
	if w.Code != http.StatusCreated || types.created == nil || types.created.CompanyID != companyA {
		t.Fatalf("code=%d created=%+v", w.Code, types.created)
	}
}

func TestUpdateCannotTargetAnotherCompany(t *testing.T) {
	r, types := newTypeRouter()
	w := do(r, "PUT", "/order-item-type?company_id="+companyA, `{"id":"`+typeOfB+`","company_id":"`+companyB+`","name":"Gilam","price":10,"unit":"area"}`)
	if w.Code != http.StatusForbidden || types.updated != nil {
		t.Fatalf("code=%d updated=%v", w.Code, types.updated)
	}
}

func TestGetByIDIsScopedToVerifiedCompany(t *testing.T) {
	r, types := newTypeRouter()
	w := do(r, "GET", "/order-item-type/get-by-primary-key/"+typeOfB, `{"company_id":"`+companyA+`"}`)
	if w.Code != http.StatusNotFound || types.gotCompany != companyA {
		t.Fatalf("code=%d gotCompany=%q body=%s", w.Code, types.gotCompany, w.Body.String())
	}
	var resp struct {
		Data json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if strings.Contains(string(resp.Data), "B's type") {
		t.Fatalf("leaked other company's type: %s", resp.Data)
	}
}

func TestDeleteIsScopedToVerifiedCompany(t *testing.T) {
	r, types := newTypeRouter()
	w := do(r, "DELETE", "/order-item-type/"+typeOfB+"?company_id="+companyA, "")
	if w.Code != http.StatusOK || types.delCompany != companyA {
		t.Fatalf("code=%d delCompany=%q", w.Code, types.delCompany)
	}
}
