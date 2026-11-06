package models

import (
	"strings"
	"testing"
)

func TestUnitIsCountable(t *testing.T) {
	if !UnitIsCountable("piece") {
		t.Fatal("piece")
	}
	for _, u := range []string{"area", "length", "weight", ""} {
		if UnitIsCountable(u) {
			t.Fatalf("%q must not be countable", u)
		}
	}
}

func TestOrderItemUnitsOrder(t *testing.T) {
	want := []string{"area", "piece", "length", "weight"}
	if strings.Join(OrderItemUnits, ",") != strings.Join(want, ",") {
		t.Fatal(OrderItemUnits)
	}
}

func TestOrderItemUnitOrderSQLListsUnitsInOrder(t *testing.T) {
	if !strings.Contains(OrderItemUnitOrderSQL, "array['area','piece','length','weight']") {
		t.Fatal(OrderItemUnitOrderSQL)
	}
}
