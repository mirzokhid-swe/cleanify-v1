package models

import (
	"fmt"
	"testing"
)

func TestDefaultOrderStatuses(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 8}
	if len(DefaultOrderStatuses) != len(numbers) {
		t.Fatalf("statuses: %d", len(DefaultOrderStatuses))
	}
	for i, s := range DefaultOrderStatuses {
		if s.Number != numbers[i] || s.Order != i+1 {
			t.Fatalf("%s: number=%d order=%d", s.Name, s.Number, s.Order)
		}
		if s.Color == "" || s.Name == "" {
			t.Fatalf("status %d: empty name or color", s.Number)
		}
		if s.Slug != fmt.Sprintf("change_status_to_%d", s.Number) {
			t.Fatalf("%s: slug=%s", s.Name, s.Slug)
		}
	}
}

func TestDefaultItemTypes(t *testing.T) {
	if len(DefaultItemTypes) != 8 {
		t.Fatalf("types: %d", len(DefaultItemTypes))
	}
	for _, it := range DefaultItemTypes {
		known := false
		for _, u := range OrderItemUnits {
			if u == it.Unit {
				known = true
			}
		}
		if !known {
			t.Fatalf("%s: unit %q", it.Name, it.Unit)
		}
		if it.Price <= 0 {
			t.Fatalf("%s: price", it.Name)
		}
	}
}

func TestDefaultRolesAndPurposes(t *testing.T) {
	names := []string{"Admin", "Menejer", "Kuryer", "Yuvuvchi"}
	if len(DefaultRoles) != len(names) {
		t.Fatalf("roles: %d", len(DefaultRoles))
	}
	for i, r := range DefaultRoles {
		if r.Name != names[i] {
			t.Fatalf("role %d: %s", i, r.Name)
		}
		if r.Name != "Admin" && len(r.Slugs) == 0 {
			t.Fatalf("%s: no slugs", r.Name)
		}
	}
	if len(DefaultPaymentPurposes) != 3 {
		t.Fatalf("purposes: %d", len(DefaultPaymentPurposes))
	}
}
