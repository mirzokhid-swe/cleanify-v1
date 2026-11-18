package postgres

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

// seedStatusOrder creates one order in demo2Company with the given items and a
// status change at day. items rows are [type, unit, width*height-as-two-values]:
// {type, unit, width, height}. Returns the order id.
func seedStatusOrder(t *testing.T, db *sqlx.DB, day time.Time, status int, userID int64, items [][4]interface{}) int {
	t.Helper()
	var orderID int
	if err := db.QueryRow(`INSERT INTO orders (company_id, status) VALUES ($1, $2) RETURNING id`, demo2Company, status).Scan(&orderID); err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if _, err := db.Exec(`INSERT INTO order_items (order_id, type, unit, price, width, height) VALUES ($1, $2, $3, 1000, $4, $5)`,
			orderID, it[0], it[1], it[2], it[3]); err != nil {
			t.Fatal(err)
		}
	}
	var user interface{}
	if userID != 0 {
		user = userID
	}
	if _, err := db.Exec(`INSERT INTO status_change_histories (historyable_type, historyable_id, status, user_id, created_at, updated_at) VALUES ('orders', $1, $2, $3::bigint, $4, $4)`,
		orderID, status, user, day); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM status_change_histories WHERE historyable_type = 'orders' AND historyable_id = $1`, orderID)
		db.Exec(`DELETE FROM order_items WHERE order_id = $1`, orderID)
		db.Exec(`DELETE FROM orders WHERE id = $1`, orderID)
	})
	return orderID
}

const metricsTestStatus = 93 // never used by real companies

func TestStatusMetricsGroupsDayUserType(t *testing.T) {
	db := testDB(t)
	repo := NewStatisticsRepo(db)
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	seedStatusOrder(t, db, today, metricsTestStatus, 90001, [][4]interface{}{
		{"13 000", "area", 3, 4},
		{"Odeyalo", "piece", 1, 1},
	})
	seedStatusOrder(t, db, today, metricsTestStatus, 90001, [][4]interface{}{{"13 000", "area", 2, 2}})
	seedStatusOrder(t, db, today, metricsTestStatus, 90002, nil)
	seedStatusOrder(t, db, yesterday, metricsTestStatus, 90001, [][4]interface{}{{"Parda", "area", 1, 2}})

	days, err := repo.GetStatusMetrics(demo2Company, metricsTestStatus, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 2 {
		t.Fatal("want 2 days, got", len(days))
	}
	d := days[0]
	if d.Date != today.Format("2006-01-02") || d.Count != 3 {
		t.Fatal(d.Date, d.Count)
	}
	if len(d.Items) != 2 || d.Items[0].Type != "13 000" || d.Items[0].Unit != "area" || d.Items[0].Qty != 16 ||
		d.Items[1].Type != "Odeyalo" || d.Items[1].Unit != "piece" || d.Items[1].Qty != 1 {
		t.Fatalf("items: %+v", d.Items)
	}
	if len(d.Users) != 2 || d.Users[0].UserID != 90001 || d.Users[0].Count != 2 ||
		d.Users[1].UserID != 90002 || d.Users[1].Count != 1 || len(d.Users[1].Items) != 0 {
		t.Fatalf("users: %+v", d.Users)
	}
	if days[1].Date != yesterday.Format("2006-01-02") || days[1].Count != 1 {
		t.Fatal(days[1].Date, days[1].Count)
	}
}

func TestStatusMetricsSameDayTwice(t *testing.T) {
	db := testDB(t)
	repo := NewStatisticsRepo(db)
	day := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 9, 0, 0, 0, time.Local)
	orderID := seedStatusOrder(t, db, day, metricsTestStatus, 90001, nil)
	if _, err := db.Exec(`INSERT INTO status_change_histories (historyable_type, historyable_id, status, user_id, created_at, updated_at) VALUES ('orders', $1, $2, 90002, $3, $3)`,
		orderID, metricsTestStatus, day.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	days, err := repo.GetStatusMetrics(demo2Company, metricsTestStatus, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 1 || days[0].Count != 1 {
		t.Fatalf("%+v", days)
	}
	if len(days[0].Users) != 1 || days[0].Users[0].UserID != 90001 {
		t.Fatalf("users: %+v", days[0].Users)
	}
}

func TestStatusMetricsWindow(t *testing.T) {
	db := testDB(t)
	repo := NewStatisticsRepo(db)
	seedStatusOrder(t, db, time.Now().AddDate(0, 0, -39), metricsTestStatus, 90001, nil)
	seedStatusOrder(t, db, time.Now().AddDate(0, 0, -40), metricsTestStatus, 90001, nil)

	days, err := repo.GetStatusMetrics(demo2Company, metricsTestStatus, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 1 {
		t.Fatalf("days=40: want 1 day, got %d", len(days))
	}
	days, err = repo.GetStatusMetrics(demo2Company, metricsTestStatus, 41)
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 2 {
		t.Fatalf("days=41: want 2 days, got %d", len(days))
	}
}

func TestStatusMetricsCompanyScope(t *testing.T) {
	db := testDB(t)
	repo := NewStatisticsRepo(db)
	seedStatusOrder(t, db, time.Now(), metricsTestStatus, 90001, nil)
	// same status in another company ("Demo")
	var otherID int
	if err := db.QueryRow(`INSERT INTO orders (company_id, status) VALUES ('8070790f-429b-449e-bfe5-fab0440c518f', $1) RETURNING id`, metricsTestStatus).Scan(&otherID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO status_change_histories (historyable_type, historyable_id, status, user_id, created_at, updated_at) VALUES ('orders', $1, $2, 90003, now(), now())`, otherID, metricsTestStatus); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM status_change_histories WHERE historyable_type = 'orders' AND historyable_id = $1`, otherID)
		db.Exec(`DELETE FROM orders WHERE id = $1`, otherID)
	})

	days, err := repo.GetStatusMetrics(demo2Company, metricsTestStatus, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 1 || days[0].Count != 1 {
		t.Fatalf("%+v", days)
	}
	if days[0].Users[0].UserID != 90001 {
		t.Fatalf("other company's user leaked: %+v", days[0].Users)
	}
}

func TestStatusMetricsEmpty(t *testing.T) {
	db := testDB(t)
	repo := NewStatisticsRepo(db)
	days, err := repo.GetStatusMetrics(demo2Company, 99, 40)
	if err != nil {
		t.Fatal(err)
	}
	if days == nil || len(days) != 0 {
		t.Fatalf("want empty non-nil slice, got %#v", days)
	}
}
