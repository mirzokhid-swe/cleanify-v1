package models

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestValidateSlots(t *testing.T) {
	if ValidateSlots([]string{"12", "14", "27"}) != nil {
		t.Fatal("valid")
	}
	if ValidateSlots([]string{}) != nil {
		t.Fatal("empty ok")
	}
	for _, bad := range [][]string{{"a1"}, {"12345678901"}, {"12", "12"}, {""}, {" 12"}} {
		if ValidateSlots(bad) == nil {
			t.Fatal(bad)
		}
	}
	many := make([]string, 21)
	for i := range many {
		many[i] = fmt.Sprint(i)
	}
	if ValidateSlots(many) == nil {
		t.Fatal("21 slots")
	}
}

func TestAskLocation(t *testing.T) {
	on := map[string]json.RawMessage{"telegram_location": json.RawMessage(`true`)}
	off := MergeSettings(nil)
	if !AskLocation(on, 1, 2) {
		t.Error("1→2 yoqiq: so'rash kerak")
	}
	if AskLocation(on, 2, 2) {
		t.Error("2→2: so'ramaslik kerak")
	}
	if AskLocation(on, 2, 3) {
		t.Error("2→3: so'ramaslik kerak")
	}
	if AskLocation(off, 1, 2) {
		t.Error("sozlama o'chiq: so'ramaslik kerak")
	}
}
