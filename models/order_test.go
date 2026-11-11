package models

import (
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
