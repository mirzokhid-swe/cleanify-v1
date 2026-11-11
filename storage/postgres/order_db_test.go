package postgres

import (
	"testing"

	"bw-erp/models"

	"github.com/lib/pq"
)

// Create writes slots (storage_slots): the INSERT column list is only
// exercised against a real database.
func TestOrderCreateWritesSlots(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	repo := NewOrderRepo(db)
	var clientID int
	if err := db.Get(&clientID, `select id from clients where company_id = $1 limit 1`, demo2Company); err != nil {
		t.Skipf("Demo 2 da mijoz yo'q: %v", err)
	}

	id, err := repo.Create(8, models.CreateOrderModel{
		CompanyID: demo2Company, ClientID: clientID, Phone: "+998900000000", Address: "test", Slots: []string{"12", "14"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer db.Exec(`delete from status_change_histories where historyable_type = 'orders' and historyable_id = $1`, id)
	defer db.Exec(`delete from orders where id = $1`, id)

	var slots pq.StringArray
	if err := db.Get(&slots, `select slots from orders where id = $1`, id); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(slots) != 2 || slots[0] != "12" || slots[1] != "14" {
		t.Fatalf("slots = %v", slots)
	}
}

// GetBusySlots: joy faqat holati 6 bo'lmagan buyurtmada band.
func TestOrderBusySlots(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	repo := NewOrderRepo(db)
	var clientID int
	if err := db.Get(&clientID, `select id from clients where company_id = $1 limit 1`, demo2Company); err != nil {
		t.Skipf("Demo 2 da mijoz yo'q: %v", err)
	}
	id, err := repo.Create(8, models.CreateOrderModel{
		CompanyID: demo2Company, ClientID: clientID, Phone: "+998900000000", Address: "test", Slots: []string{"97", "98"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer db.Exec(`delete from status_change_histories where historyable_type = 'orders' and historyable_id = $1`, id)
	defer db.Exec(`delete from orders where id = $1`, id)

	has := func(slot string) bool {
		busy, err := repo.GetBusySlots(demo2Company)
		if err != nil {
			t.Fatalf("busy: %v", err)
		}
		for _, b := range busy {
			if b.OrderID == id && b.Slot == slot {
				return true
			}
		}
		return false
	}
	if !has("97") || !has("98") {
		t.Fatal("yangi buyurtma joylari band bo'lishi kerak")
	}
	if _, err := db.Exec(`update orders set status = 6 where id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if has("97") {
		t.Fatal("topshirilgan buyurtma joyi band bo'lmasligi kerak")
	}
}
