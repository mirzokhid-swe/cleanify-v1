package postgres

import (
	"sync"
	"testing"

	"bw-erp/models"
)

// Two identical payments sent at the same time (double tap) must store one row:
// the duplicate rule and the remaining-amount check run under a row lock.
func TestAddPaymentConcurrentDuplicateStoresOnce(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	repo := &orderRepo{db: db}
	const qor = "82e1054c-48ef-4ed5-9e6b-9d3fbbf00477"
	const orderID = 43008
	db.Exec(`update transactions set deleted_at = now(), deleted_by = 8 where payer_type = 'orders' and payer_id = $1 and deleted_at is null`, orderID)
	defer db.Exec(`update transactions set deleted_at = now(), deleted_by = 8 where payer_type = 'orders' and payer_id = $1 and deleted_at is null`, orderID)
	before, err := repo.GetPaymentSummary(orderID)
	if err != nil || before.Remaining <= 0 {
		t.Fatalf("fixture: %v %+v", err, before)
	}
	req := models.AddOrderPaymentRequest{CompanyID: qor, OrderID: orderID, Amount: before.Remaining, PaymentType: 1}
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func(i int) { defer wg.Done(); _, errs[i] = repo.AddPayment(8, req) }(i)
	}
	wg.Wait()
	var stored int
	db.Get(&stored, `select count(*) from transactions where payer_type = 'orders' and payer_id = $1 and deleted_at is null`, orderID)
	after, _ := repo.GetPaymentSummary(orderID)
	if stored != 1 || after.Paid != before.Remaining {
		t.Fatalf("stored %d rows, paid %v of %v (errs %v)", stored, after.Paid, before.Remaining, errs)
	}
}
