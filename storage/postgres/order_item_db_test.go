package postgres

import (
	"testing"

	"bw-erp/models"

	"github.com/lib/pq"
)

// demoOrderAndType: Demo 2 kompaniyasining bitta buyurtmasi va turi (test ma'lumoti).
func demoOrderAndType(t *testing.T, db interface {
	Get(dest interface{}, query string, args ...interface{}) error
}) (int, string) {
	var orderID int
	var typeID string
	if err := db.Get(&orderID, `select id from orders where company_id = $1 order by id desc limit 1`, demo2Company); err != nil {
		t.Skipf("Demo 2 da buyurtma yo'q: %v", err)
	}
	if err := db.Get(&typeID, `select id from order_item_types where company_id = $1 limit 1`, demo2Company); err != nil {
		t.Skipf("Demo 2 da tur yo'q: %v", err)
	}
	return orderID, typeID
}

func TestCreateWritesCountRowsInOneTransaction(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	repo := NewOrderItemRepo(db)
	orderID, typeID := demoOrderAndType(t, db)

	ids, err := repo.Create(8, models.CreateOrderItemModel{
		OrderID: orderID, OrderItemTypeID: typeID, Price: 25000, Width: 3, Height: 1,
		Count: 3, ItemType: "PW metr", Unit: "length",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() {
		for _, id := range ids {
			repo.DeleteByID(id)
			db.Exec(`delete from status_change_histories where historyable_type = 'order_items' and historyable_id = $1`, id)
		}
	}()
	if len(ids) != 3 {
		t.Fatalf("ids %v", ids)
	}
	var n, histories int
	db.Get(&n, `select count(*) from order_items where id = any($1) and unit = 'length' and not is_countable and width = 3 and height = 1`, pq.Array(ids))
	db.Get(&histories, `select count(*) from status_change_histories where historyable_type = 'order_items' and historyable_id = any($1)`, pq.Array(ids))
	if n != 3 || histories != 3 {
		t.Fatalf("rows %d histories %d", n, histories)
	}
}

func TestCreateRollsBackOnFailure(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	repo := NewOrderItemRepo(db)
	orderID, _ := demoOrderAndType(t, db)

	var before, after int
	db.Get(&before, `select count(*) from order_items where order_id = $1`, orderID)
	// mavjud bo'lmagan tur — FK xatosi
	_, err := repo.Create(8, models.CreateOrderItemModel{
		OrderID: orderID, OrderItemTypeID: "00000000-0000-0000-0000-000000000001", Price: 1, Width: 1, Height: 1,
		Count: 2, ItemType: "x", Unit: "area",
	})
	db.Get(&after, `select count(*) from order_items where order_id = $1`, orderID)
	if err == nil || after != before {
		t.Fatalf("err=%v rows before %d after %d", err, before, after)
	}
}

func TestUpdateWritesUnitAndType(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	repo := NewOrderItemRepo(db)
	orderID, typeID := demoOrderAndType(t, db)

	ids, err := repo.Create(8, models.CreateOrderItemModel{
		OrderID: orderID, OrderItemTypeID: typeID, Price: 13000, Width: 3, Height: 4, ItemType: "PW m2", Unit: "area",
	})
	if err != nil || len(ids) != 1 {
		t.Fatalf("create: %v %v", ids, err)
	}
	defer func() {
		repo.DeleteByID(ids[0])
		db.Exec(`delete from status_change_histories where historyable_type = 'order_items' and historyable_id = $1`, ids[0])
	}()

	_, err = repo.Update(models.UpdateOrderItemRequest{
		ID: ids[0], OrderItemTypeID: typeID, Price: 25000, Width: 5, Height: 1, ItemType: "PW metr", Unit: "length",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	var row struct {
		Unit        string  `db:"unit"`
		Type        string  `db:"type"`
		Width       float32 `db:"width"`
		Height      float32 `db:"height"`
		Price       float32 `db:"price"`
		IsCountable bool    `db:"is_countable"`
	}
	if err := db.Get(&row, `select unit, type, width, height, price, is_countable from order_items where id = $1`, ids[0]); err != nil {
		t.Fatal(err)
	}
	if row.Unit != "length" || row.Type != "PW metr" || row.Width != 5 || row.Height != 1 || row.Price != 25000 || row.IsCountable {
		t.Fatalf("row %+v", row)
	}
}
