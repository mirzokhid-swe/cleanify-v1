package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// /orders/busy-slots statik yo'l /orders/:order-id dan oldin tanlanishi kerak.
func TestBusySlotsRouteBeatsOrderIDParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/orders/:order-id", func(c *gin.Context) { c.String(200, "param") })
	r.GET("/orders/busy-slots", func(c *gin.Context) { c.String(200, "static") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/orders/busy-slots", nil))
	if w.Body.String() != "static" {
		t.Fatalf("got %q", w.Body.String())
	}
}
