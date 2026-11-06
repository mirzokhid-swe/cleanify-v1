package models

import (
	"encoding/json"
	"testing"
)

func TestDefaultSettingsHasNineKeys(t *testing.T) {
	d := DefaultSettings()
	if len(d) != 9 {
		t.Fatal(len(d))
	}
	if string(d["permissions_mode"]) != `"individual"` {
		t.Fatal(string(d["permissions_mode"]))
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
		{"permissions_mode", `"role"`, `"role"`},
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
	if string(m["kpi_enabled"]) != "true" || string(m["volume_mode"]) != `"type"` || len(m) != 9 {
		t.Fatal(m)
	}
	if len(MergeSettings(nil)) != 9 {
		t.Fatal("nil rows must give defaults")
	}
}
