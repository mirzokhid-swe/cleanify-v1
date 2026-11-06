package postgres

import (
	"strings"
	"testing"
	"time"

	"bw-erp/models"
)

func TestOrdersListFilterIgnoresShortSearch(t *testing.T) {
	f, p := ordersListFilter("c", models.OrdersListRequest{Search: "ab"})
	if strings.Contains(f, "ILIKE") {
		t.Fatal(f)
	}
	if _, ok := p["search"]; ok {
		t.Fatal("search bound")
	}
	if p["company_id"] != "c" || !strings.Contains(f, "o.company_id = :company_id") {
		t.Fatal(f, p)
	}
}

func TestOrdersListFilterWrapsSearch(t *testing.T) {
	f, p := ordersListFilter("c", models.OrdersListRequest{Search: "ming"})
	if !strings.Contains(f, "ILIKE :search") || p["search"] != "%ming%" {
		t.Fatal(f, p)
	}
}

func TestOrdersListFilterDateToInclusive(t *testing.T) {
	d, _ := time.Parse("2006-01-02", "2026-11-03")
	f, p := ordersListFilter("c", models.OrdersListRequest{DateTo: d, DateFrom: d})
	if !strings.Contains(f, "o.created_at < (:date_to::date + 1)") || !strings.Contains(f, "o.created_at >= :date_from::date") {
		t.Fatal(f)
	}
	if p["date_to"] != d || p["date_from"] != d {
		t.Fatal(p)
	}
}

func TestOrdersListDataFilterDebtor(t *testing.T) {
	if !strings.Contains(ordersListDataFilter(models.OrdersListRequest{Debtor: true}), "status = 6 AND payment_status IN (1,2)") {
		t.Fatal("debtor")
	}
	if ordersListDataFilter(models.OrdersListRequest{}) != "" {
		t.Fatal("empty")
	}
	if !strings.Contains(ordersListDataFilter(models.OrdersListRequest{Status: 3}), "status = :status") {
		t.Fatal("status")
	}
}

func TestOrdersListLimit(t *testing.T) {
	for in, want := range map[int32]int32{0: 150, 20: 20, 151: 150, -5: 150} {
		if got := ordersListLimit(in); got != want {
			t.Fatalf("limit %d: got %d want %d", in, got, want)
		}
	}
}
