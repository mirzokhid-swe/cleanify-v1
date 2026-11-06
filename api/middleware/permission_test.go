package middleware

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"bw-erp/models"

	"github.com/gin-gonic/gin"
)

func TestHasPermission(t *testing.T) {
	if !hasPermission("get_orders_list|edit_order", "edit_order") {
		t.Fatal("expected true")
	}
	if hasPermission("get_orders_list|edit_order_item", "edit_order") {
		t.Fatal("prefix must not match")
	}
	if hasPermission("", "edit_order") {
		t.Fatal("empty can")
	}
}

func TestCompanyIDFromRequestReadsBodyAndRestoresIt(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/x", strings.NewReader(`{"company_id":"c-1","name":"Gilam"}`))
	if got := CompanyIDFromRequest(c); got != "c-1" {
		t.Fatalf("got %q", got)
	}
	b, _ := io.ReadAll(c.Request.Body)
	if !strings.Contains(string(b), `"name":"Gilam"`) {
		t.Fatal("body not restored")
	}
}

func TestCompanyIDFromRequestPrefersQuery(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x?company_id=c-9", nil)
	if got := CompanyIDFromRequest(c); got != "c-9" {
		t.Fatalf("got %q", got)
	}
}

func TestRequirePermissionDeniesWithoutCompany(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x", nil)
	c.Set("user", models.User{UserPermissionByCompany: []models.UserPermissionByCompany{{CompanyID: "c-1", Can: "get_order_item_type"}}})
	RequirePermission("get_order_item_type")(c)
	if !c.IsAborted() || w.Code != 403 {
		t.Fatalf("aborted=%v code=%d", c.IsAborted(), w.Code)
	}
}

func TestRequirePermissionDeniesWrongCompanyOrSlug(t *testing.T) {
	for name, tc := range map[string]struct{ company, can string }{
		"other company": {"c-2", "get_order_item_type"},
		"missing slug":  {"c-1", "get_orders_list"},
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/x?company_id=c-1", nil)
		c.Set("user", models.User{UserPermissionByCompany: []models.UserPermissionByCompany{{CompanyID: tc.company, Can: tc.can}}})
		RequirePermission("get_order_item_type")(c)
		if !c.IsAborted() || w.Code != 403 {
			t.Fatalf("%s: aborted=%v code=%d", name, c.IsAborted(), w.Code)
		}
	}
}

func TestRequirePermissionDeniesWithoutUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x?company_id=c-1", nil)
	RequirePermission("get_order_item_type")(c)
	if !c.IsAborted() || w.Code != 403 {
		t.Fatalf("aborted=%v code=%d", c.IsAborted(), w.Code)
	}
}

func TestRequirePermissionAllows(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x?company_id=c-1", nil)
	c.Set("user", models.User{UserPermissionByCompany: []models.UserPermissionByCompany{{CompanyID: "c-1", Can: "create_order_item_type|get_order_item_type"}}})
	RequirePermission("get_order_item_type")(c)
	if c.IsAborted() {
		t.Fatalf("should not abort, code=%d", w.Code)
	}
	if got := c.GetString("company_id"); got != "c-1" {
		t.Fatalf("company_id in context: %q", got)
	}
}

func TestRequirePermissionDeniesQueryBodyMismatch(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/x?company_id=c-1", strings.NewReader(`{"company_id":"c-2","name":"Gilam"}`))
	c.Set("user", models.User{UserPermissionByCompany: []models.UserPermissionByCompany{{CompanyID: "c-1", Can: "create_order_item_type"}}})
	RequirePermission("create_order_item_type")(c)
	if !c.IsAborted() || w.Code != 403 {
		t.Fatalf("aborted=%v code=%d", c.IsAborted(), w.Code)
	}
}

func TestRequirePermissionResponseIsDistinguishableFromAuthFailure(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x?company_id=c-1", nil)
	c.Set("user", models.User{UserPermissionByCompany: []models.UserPermissionByCompany{{CompanyID: "c-1", Can: "get_orders_list"}}})
	RequirePermission("get_order_item_type")(c)
	if !strings.Contains(w.Body.String(), `"status":"PERMISSION_DENIED"`) {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestRequirePermissionBlockedCompanyIs423(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x?company_id=c-1", nil)
	c.Set("user", models.User{UserPermissionByCompany: []models.UserPermissionByCompany{{CompanyID: "c-1", Can: "get_orders_list", IsBlocked: true}}})
	RequirePermission("get_orders_list")(c)
	if !c.IsAborted() || w.Code != 423 {
		t.Fatalf("aborted=%v code=%d", c.IsAborted(), w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status":"LOCKED"`) {
		t.Fatalf("body %s", w.Body.String())
	}
}
