package models

import "testing"

func TestMeasureForUnitArea(t *testing.T) {
	w, h, err := MeasureForUnit("area", 3, 4)
	if err != nil || w != 3 || h != 4 {
		t.Fatalf("got %v %v %v", w, h, err)
	}
	if _, _, err := MeasureForUnit("area", 0, 4); err != ErrAreaRequired {
		t.Fatalf("width 0: %v", err)
	}
	if _, _, err := MeasureForUnit("area", 3, 0); err != ErrAreaRequired {
		t.Fatalf("height 0: %v", err)
	}
}

func TestMeasureForUnitPieceIgnoresSizes(t *testing.T) {
	w, h, err := MeasureForUnit("piece", 7, 9)
	if err != nil || w != 1 || h != 1 {
		t.Fatalf("got %v %v %v", w, h, err)
	}
}

// Metr va kg: miqdor width da, height = 1 (egasi, 2026-11-09).
func TestMeasureForUnitLengthAndWeightUseWidth(t *testing.T) {
	if w, h, err := MeasureForUnit("length", 3, 0); err != nil || w != 3 || h != 1 {
		t.Fatalf("length: %v %v %v", w, h, err)
	}
	if w, h, err := MeasureForUnit("weight", 4.5, 0); err != nil || w != 4.5 || h != 1 {
		t.Fatalf("weight: %v %v %v", w, h, err)
	}
	if _, _, err := MeasureForUnit("length", 0, 0); err != ErrQuantityRequired {
		t.Fatalf("0/0: %v", err)
	}
}

// Eski front ikkala tomonni yuboradi: 3×1 ham, 1×3 ham 3 m.
func TestMeasureForUnitLengthAcceptsEitherSide(t *testing.T) {
	cases := []struct{ w, h, want float32 }{{3, 1, 3}, {1, 3, 3}, {0, 3, 3}, {2, 3, 6}}
	for _, c := range cases {
		w, h, err := MeasureForUnit("length", c.w, c.h)
		if err != nil || w != c.want || h != 1 {
			t.Fatalf("(%v,%v) → %v %v %v, want %v 1", c.w, c.h, w, h, err, c.want)
		}
	}
}

func TestMeasureForUnitUnknownIsArea(t *testing.T) {
	if _, _, err := MeasureForUnit("", 0, 0); err != ErrAreaRequired {
		t.Fatalf("got %v", err)
	}
}

func TestRowsDefaultsToOne(t *testing.T) {
	if (CreateOrderItemModel{}).Rows() != 1 {
		t.Fatal("count 0 → 1")
	}
	if (CreateOrderItemModel{Count: 3}).Rows() != 3 {
		t.Fatal("count 3 → 3")
	}
}
