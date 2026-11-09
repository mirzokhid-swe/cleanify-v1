package models

import (
	"encoding/json"
	"testing"
)

func TestDefaultSettingsHasTenKeys(t *testing.T) {
	d := DefaultSettings()
	if len(d) != 10 {
		t.Fatal(len(d))
	}
	if string(d["payment_tolerance"]) != "0" || string(d["volume_mode"]) != `"type"` || string(d["kpi_enabled"]) != "false" {
		t.Fatal(d)
	}
}

func TestValidateSetting(t *testing.T) {
	ok := []struct{ k, v, want string }{
		{"payment_tolerance", "15000", "15000"},
		{"payment_tolerance", " 0 ", "0"},
		{"kpi_enabled", "true", "true"},
		{"volume_mode", `"worker"`, `"worker"`},
		{"item_count_enabled", "false", "false"},
		{"item_count_max", "1", "1"},
		{"item_count_max", "100", "100"},
	}
	for _, c := range ok {
		got, err := ValidateSetting(c.k, json.RawMessage(c.v))
		if err != nil {
			t.Fatal(c.k, err)
		}
		if string(got) != c.want {
			t.Fatalf("%s: got %s want %s", c.k, got, c.want)
		}
	}
	bad := []struct{ k, v string }{
		{"payment_tolerance", "-1"},
		{"payment_tolerance", "15.5"},
		{"payment_tolerance", `"15000"`},
		{"kpi_enabled", `"true"`},
		{"kpi_enabled", "1"},
		{"volume_mode", `"x"`},
		{"volume_mode", "true"},
		{"unknown", "1"},
		{"kpi_enabled", ""},
		{"item_count_max", "0"},
		{"item_count_max", "101"},
	}
	for _, c := range bad {
		if _, err := ValidateSetting(c.k, json.RawMessage(c.v)); err == nil {
			t.Fatal("expected error", c.k, c.v)
		}
	}
	if _, err := ValidateSetting("unknown", json.RawMessage("1")); err != ErrUnknownSetting {
		t.Fatalf("want ErrUnknownSetting, got %v", err)
	}
}

func TestMergeSettingsOverridesDefaults(t *testing.T) {
	m := MergeSettings(map[string]json.RawMessage{"kpi_enabled": json.RawMessage("true")})
	if string(m["kpi_enabled"]) != "true" || string(m["volume_mode"]) != `"type"` || len(m) != 10 {
		t.Fatal(m)
	}
	if len(MergeSettings(nil)) != 10 {
		t.Fatal("nil rows must give defaults")
	}
}

func TestItemCountMax(t *testing.T) {
	cases := []struct {
		rows map[string]json.RawMessage
		want int
	}{
		{nil, 20}, // standart: "Soni" ko'rinadi, ko'pi bilan 20
		{map[string]json.RawMessage{"item_count_max": json.RawMessage("50")}, 50},
		{map[string]json.RawMessage{"item_count_enabled": json.RawMessage("false"), "item_count_max": json.RawMessage("50")}, 1},
	}
	for _, c := range cases {
		if got := ItemCountMax(MergeSettings(c.rows)); got != c.want {
			t.Fatalf("%v: got %d want %d", c.rows, got, c.want)
		}
	}
}

func TestSettingsCatalogHasNoPermissionsMode(t *testing.T) {
	if _, ok := SettingsCatalog["permissions_mode"]; ok {
		t.Fatal("permissions_mode must be removed")
	}
}
