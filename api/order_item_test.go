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

// Orders: 1 belongs to company A, 2 to company B.
// Items: 10 is on order 1, 20 is on order 2.

type itemsOrderRepo struct {
	repo.OrderI
}

func (itemsOrderRepo) GetByPrimaryKey(id int) (models.OrderShowResponse, error) {
	switch id {
	case 1:
		return models.OrderShowResponse{Order: models.Order{ID: 1, CompanyID: companyA}}, nil
	case 2:
		return models.OrderShowResponse{Order: models.Order{ID: 2, CompanyID: companyB}}, nil
	}
	return models.OrderShowResponse{}, sql.ErrNoRows
}

type itemsItemRepo struct {
	repo.OrderItemI
	created       *models.CreateOrderItemModel
	updated       *models.UpdateOrderItemRequest
	statusChanged *models.UpdateOrderItemStatusRequest
	deleted       int
}

func (f *itemsItemRepo) GetOrderID(itemID int) (int, error) {
	switch itemID {
	case 10:
		return 1, nil
	case 20:
		return 2, nil
	}
	return 0, sql.ErrNoRows
}

func (f *itemsItemRepo) Create(userID int64, e models.CreateOrderItemModel) ([]int, error) {
	f.created = &e
	ids := []int{}
	for i := 0; i < e.Rows(); i++ {
		ids = append(ids, 100+i)
	}
	return ids, nil
}

func (f *itemsItemRepo) Update(e models.UpdateOrderItemRequest) (int64, error) {
	f.updated = &e
	return 1, nil
}

func (f *itemsItemRepo) DeleteByID(id int) error {
	f.deleted = id
	return nil
}

func (f *itemsItemRepo) UpdateStatus(userID int64, e models.UpdateOrderItemStatusRequest) (int64, error) {
	f.statusChanged = &e
	return 1, nil
}

// Only company A has these types, one per unit.
type itemsTypeRepo struct {
	repo.OrderItemTypeI
}

const (
	typeOfA   = "44444444-4444-4444-4444-444444444444" // area
	pieceOfA  = "55555555-5555-5555-5555-555555555555"
	lengthOfA = "66666666-6666-6666-6666-666666666666"
	weightOfA = "77777777-7777-7777-7777-777777777777"
)

func (itemsTypeRepo) GetById(id string, companyID string) (models.OrderItemByCompany, error) {
	if companyID != companyA {
		return models.OrderItemByCompany{}, sql.ErrNoRows
	}
	switch id {
	case typeOfA:
		return models.OrderItemByCompany{ID: id, Name: "Gilam", Unit: "area", CopmanyID: companyA}, nil
	case pieceOfA:
		return models.OrderItemByCompany{ID: id, Name: "Odeyalo", Unit: "piece", IsCountable: true, CopmanyID: companyA}, nil
	case lengthOfA:
		return models.OrderItemByCompany{ID: id, Name: "Ko'rpacha", Unit: "length", CopmanyID: companyA}, nil
	case weightOfA:
		return models.OrderItemByCompany{ID: id, Name: "Parda", Unit: "weight", CopmanyID: companyA}, nil
	}
	return models.OrderItemByCompany{}, sql.ErrNoRows
}

type itemsStorage struct {
	storage.StorageI
	items     *itemsItemRepo
	companies *fakeCompanies
}

func (s itemsStorage) Order() repo.OrderI                 { return itemsOrderRepo{} }
func (s itemsStorage) OrderItem() repo.OrderItemI         { return s.items }
func (s itemsStorage) OrderItemType() repo.OrderItemTypeI { return itemsTypeRepo{} }
func (s itemsStorage) Company() repo.CompanyStorageI      { return s.companies }

// itemUser works in company A with every item permission, nothing in B.
func itemUser(c *gin.Context) {
	c.Set("user", models.User{ID: 7, UserPermissionByCompany: []models.UserPermissionByCompany{{
		CompanyID: companyA,
		Can:       "create_order_item|edit_order_item|delete_order_item",
	}}})
}

// itemUserNoPerms works in company A but has no item permission.
func itemUserNoPerms(c *gin.Context) {
	c.Set("user", models.User{ID: 7, UserPermissionByCompany: []models.UserPermissionByCompany{{
		CompanyID: companyA,
		Can:       "get_orders_list",
	}}})
}

func newItemRouter(user gin.HandlerFunc) (*gin.Engine, *itemsItemRepo) {
	gin.SetMode(gin.TestMode)
	items := &itemsItemRepo{}
	h := handlers.Handler{Stg: itemsStorage{items: items, companies: &fakeCompanies{}}}
	r := gin.New()
	orderItemRoutes(r.Group("/order-items"), &h, user)
	return r, items
}

// newItemRouterWithSettings: company A's settings are the given rows.
func newItemRouterWithSettings(settings map[string]json.RawMessage) (*gin.Engine, *itemsItemRepo) {
	gin.SetMode(gin.TestMode)
	items := &itemsItemRepo{}
	h := handlers.Handler{Stg: itemsStorage{items: items, companies: &fakeCompanies{settings: settings}}}
	r := gin.New()
	orderItemRoutes(r.Group("/order-items"), &h, itemUser)
	return r, items
}

func responseStatus(body []byte) string {
	var resp struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(body, &resp)
	return resp.Status
}

func TestCreateItemOnForeignOrderIsDenied(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":2,"order_item_type_id":"`+typeOfA+`","price":1,"width":1,"height":1}`)
	if w.Code != http.StatusForbidden || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" || items.created != nil {
		t.Fatalf("code=%d body=%s created=%v", w.Code, w.Body.String(), items.created)
	}
}

func TestCreateItemWithoutSlugIsDenied(t *testing.T) {
	r, items := newItemRouter(itemUserNoPerms)
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+typeOfA+`","price":1,"width":1,"height":1}`)
	if w.Code != http.StatusForbidden || items.created != nil {
		t.Fatalf("code=%d created=%v", w.Code, items.created)
	}
}

func TestCreateItemAllowed(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+typeOfA+`","price":1,"width":1,"height":1}`)
	if w.Code != http.StatusCreated || items.created == nil || items.created.OrderID != 1 {
		t.Fatalf("code=%d body=%s created=%v", w.Code, w.Body.String(), items.created)
	}
}

func TestCreateItemUnknownOrderIs404(t *testing.T) {
	r, _ := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":999,"order_item_type_id":"`+typeOfA+`","price":1,"width":1,"height":1}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestCreateItemWithTypeOfOtherCompanyIs400(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+typeOfB+`","price":1,"width":1,"height":1}`)
	if w.Code != http.StatusBadRequest || items.created != nil {
		t.Fatalf("code=%d created=%v", w.Code, items.created)
	}
}

func TestEditDeleteStatusOnForeignItemDenied(t *testing.T) {
	r, items := newItemRouter(itemUser)
	if w := do(r, "POST", "/order-items/edit", `{"id":20,"order_item_type_id":"`+typeOfA+`","price":1,"width":1,"height":1}`); w.Code != http.StatusForbidden {
		t.Fatalf("edit code=%d", w.Code)
	}
	if w := do(r, "DELETE", "/order-items/20", ""); w.Code != http.StatusForbidden {
		t.Fatalf("delete code=%d", w.Code)
	}
	if w := do(r, "POST", "/order-items/edit-status", `{"id":20,"status":1}`); w.Code != http.StatusForbidden {
		t.Fatalf("edit-status code=%d", w.Code)
	}
	if items.updated != nil || items.deleted != 0 || items.statusChanged != nil {
		t.Fatalf("foreign item was changed: updated=%v deleted=%d status=%v", items.updated, items.deleted, items.statusChanged)
	}
}

func TestEditDeleteStatusWithoutSlugDenied(t *testing.T) {
	r, items := newItemRouter(itemUserNoPerms)
	do(r, "POST", "/order-items/edit", `{"id":10,"order_item_type_id":"`+typeOfA+`","price":1,"width":1,"height":1}`)
	do(r, "DELETE", "/order-items/10", "")
	do(r, "POST", "/order-items/edit-status", `{"id":10,"status":1}`)
	if items.updated != nil || items.deleted != 0 || items.statusChanged != nil {
		t.Fatalf("item changed without permission: updated=%v deleted=%d status=%v", items.updated, items.deleted, items.statusChanged)
	}
}

func TestEditDeleteStatusOwnItemAllowed(t *testing.T) {
	r, items := newItemRouter(itemUser)
	if w := do(r, "POST", "/order-items/edit", `{"id":10,"order_item_type_id":"`+typeOfA+`","price":1,"width":1,"height":1}`); w.Code != http.StatusOK || items.updated == nil {
		t.Fatalf("edit code=%d body=%s", w.Code, w.Body.String())
	}
	if w := do(r, "POST", "/order-items/edit-status", `{"id":10,"status":1}`); w.Code != http.StatusOK || items.statusChanged == nil {
		t.Fatalf("edit-status code=%d body=%s", w.Code, w.Body.String())
	}
	if w := do(r, "DELETE", "/order-items/10", ""); w.Code != http.StatusOK || items.deleted != 10 {
		t.Fatalf("delete code=%d deleted=%d", w.Code, items.deleted)
	}
}

func TestUnknownItemIs404(t *testing.T) {
	r, _ := newItemRouter(itemUser)
	if w := do(r, "DELETE", "/order-items/999", ""); w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestCreateAreaWithoutSizesIs400(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+typeOfA+`","price":1,"width":0,"height":0}`)
	if w.Code != http.StatusBadRequest || items.created != nil {
		t.Fatalf("code=%d created=%v", w.Code, items.created)
	}
}

func TestCreatePieceForcesOneByOneAndCountRows(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+pieceOfA+`","price":50000,"width":7,"height":9,"count":3}`)
	if w.Code != http.StatusCreated || items.created == nil {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	c := items.created
	if c.Width != 1 || c.Height != 1 || c.Unit != "piece" || !c.IsCountable || c.ItemType != "Odeyalo" {
		t.Fatalf("created %+v", c)
	}
	var resp struct {
		Data struct {
			IDs []int `json:"ids"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data.IDs) != 3 {
		t.Fatalf("ids %s", w.Body.String())
	}
}

// Metr/kg: miqdor width da, height = 1 (egasi, 2026-11-09).
func TestCreateLengthStoresQuantityInWidth(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+lengthOfA+`","price":25000,"width":3}`)
	if w.Code != http.StatusCreated || items.created == nil {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	c := items.created
	if c.Width != 3 || c.Height != 1 || c.Unit != "length" || c.IsCountable {
		t.Fatalf("created %+v", c)
	}
}

func TestCreateLengthWithoutQuantityIs400(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+lengthOfA+`","price":25000}`)
	if w.Code != http.StatusBadRequest || items.created != nil {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestCreateCountAboveLimitIs400(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+pieceOfA+`","price":1,"count":21}`)
	if w.Code != http.StatusBadRequest || items.created != nil {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestCreateCountUpToCompanyMax(t *testing.T) {
	r, items := newItemRouterWithSettings(map[string]json.RawMessage{"item_count_max": json.RawMessage("50")})
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+pieceOfA+`","price":1,"count":30}`)
	if w.Code != http.StatusCreated || items.created == nil || items.created.Rows() != 30 {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestCreateCountAboveCompanyMaxIs400(t *testing.T) {
	r, items := newItemRouterWithSettings(map[string]json.RawMessage{"item_count_max": json.RawMessage("5")})
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+pieceOfA+`","price":1,"count":6}`)
	if w.Code != http.StatusBadRequest || items.created != nil {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestCreateCountWhenDisabledIs400(t *testing.T) {
	r, items := newItemRouterWithSettings(map[string]json.RawMessage{"item_count_enabled": json.RawMessage("false")})
	w := do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+pieceOfA+`","price":1,"count":2}`)
	if w.Code != http.StatusBadRequest || items.created != nil {
		t.Fatalf("code=%d", w.Code)
	}
	// "Soni" yo'q bo'lsa ham bitta buyum qo'shiladi.
	w = do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+pieceOfA+`","price":1}`)
	if w.Code != http.StatusCreated || items.created == nil || items.created.Rows() != 1 {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestClientCannotSetUnitOrCountable(t *testing.T) {
	r, items := newItemRouter(itemUser)
	do(r, "POST", "/order-items", `{"order_id":1,"order_item_type_id":"`+typeOfA+`","price":1,"width":2,"height":3,"unit":"piece","is_countable":true,"item_type":"x"}`)
	if items.created == nil || items.created.Unit != "area" || items.created.IsCountable || items.created.ItemType != "Gilam" {
		t.Fatalf("created %+v", items.created)
	}
}

func TestEditNormalisesByUnit(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items/edit", `{"id":10,"order_item_type_id":"`+weightOfA+`","price":30000,"width":4.5,"height":0}`)
	if w.Code != http.StatusOK || items.updated == nil {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
	u := items.updated
	if u.Width != 4.5 || u.Height != 1 || u.Unit != "weight" || u.ItemType != "Parda" || u.OrderItemTypeID != weightOfA {
		t.Fatalf("updated %+v", u)
	}
}

func TestEditAreaWithoutSizesIs400(t *testing.T) {
	r, items := newItemRouter(itemUser)
	w := do(r, "POST", "/order-items/edit", `{"id":10,"order_item_type_id":"`+typeOfA+`","price":1,"width":0,"height":2}`)
	if w.Code != http.StatusBadRequest || items.updated != nil {
		t.Fatalf("code=%d", w.Code)
	}
}
