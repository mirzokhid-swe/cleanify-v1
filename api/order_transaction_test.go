package api

import (
	"database/sql"
	"net/http"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/models"
	"bw-erp/storage"
	"bw-erp/storage/repo"

	"github.com/gin-gonic/gin"
)

// Payments: 1 entered by user 7 (company A), 2 entered by user 9 (company A), 3 entered by user 7 (company B).
type txOrderRepo struct {
	repo.OrderI
	deleted int
}

func (f *txOrderRepo) GetOrderPayment(id int) (models.OrderPaymentRef, error) {
	switch id {
	case 1:
		return models.OrderPaymentRef{CompanyID: companyA, ReceiverID: 7}, nil
	case 2:
		return models.OrderPaymentRef{CompanyID: companyA, ReceiverID: 9}, nil
	case 3:
		return models.OrderPaymentRef{CompanyID: companyB, ReceiverID: 7}, nil
	}
	return models.OrderPaymentRef{}, sql.ErrNoRows
}

func (f *txOrderRepo) DeleteTransaction(userID int64, id int, companyID string) (models.PaymentSummary, error) {
	f.deleted = id
	return models.PaymentSummary{}, nil
}

type txStorage struct {
	storage.StorageI
	orders *txOrderRepo
}

func (s txStorage) Order() repo.OrderI { return s.orders }

func txUser(can string, blocked bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", models.User{ID: 7, UserPermissionByCompany: []models.UserPermissionByCompany{{
			CompanyID: companyA, Can: can, IsBlocked: blocked,
		}}})
	}
}

func newTxRouter(user gin.HandlerFunc) (*gin.Engine, *txOrderRepo) {
	gin.SetMode(gin.TestMode)
	orders := &txOrderRepo{}
	h := handlers.Handler{Stg: txStorage{orders: orders}}
	r := gin.New()
	r.DELETE("/orders/transactions/:id", user, h.DeleteOrderTransaction)
	return r, orders
}

func TestDeleteAnyPaymentWithPermission(t *testing.T) {
	r, orders := newTxRouter(txUser("delete_transaction", false))
	w := do(r, "DELETE", "/orders/transactions/2?company_id="+companyA, "")
	if w.Code != http.StatusOK || orders.deleted != 2 {
		t.Fatalf("code=%d deleted=%d", w.Code, orders.deleted)
	}
}

func TestDeleteOwnPaymentWithoutPermission(t *testing.T) {
	r, orders := newTxRouter(txUser("add_payment", false))
	w := do(r, "DELETE", "/orders/transactions/1?company_id="+companyA, "")
	if w.Code != http.StatusOK || orders.deleted != 1 {
		t.Fatalf("code=%d deleted=%d body=%s", w.Code, orders.deleted, w.Body.String())
	}
}

func TestDeleteOthersPaymentWithoutPermissionDenied(t *testing.T) {
	r, orders := newTxRouter(txUser("add_payment", false))
	w := do(r, "DELETE", "/orders/transactions/2?company_id="+companyA, "")
	if w.Code != http.StatusForbidden || orders.deleted != 0 || responseStatus(w.Body.Bytes()) != "PERMISSION_DENIED" {
		t.Fatalf("code=%d deleted=%d", w.Code, orders.deleted)
	}
}

func TestDeleteOwnPaymentInForeignCompanyDenied(t *testing.T) {
	r, orders := newTxRouter(txUser("delete_transaction", false))
	w := do(r, "DELETE", "/orders/transactions/3?company_id="+companyB, "")
	if w.Code != http.StatusForbidden || orders.deleted != 0 {
		t.Fatalf("code=%d deleted=%d", w.Code, orders.deleted)
	}
}

func TestDeleteOwnPaymentInBlockedCompanyDenied(t *testing.T) {
	r, orders := newTxRouter(txUser("", true))
	w := do(r, "DELETE", "/orders/transactions/1?company_id="+companyA, "")
	if w.Code != http.StatusForbidden || orders.deleted != 0 {
		t.Fatalf("code=%d deleted=%d", w.Code, orders.deleted)
	}
}

func TestDeleteUnknownPaymentIs404(t *testing.T) {
	r, _ := newTxRouter(txUser("delete_transaction", false))
	if w := do(r, "DELETE", "/orders/transactions/99?company_id="+companyA, ""); w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}
