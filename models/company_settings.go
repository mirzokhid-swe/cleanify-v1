package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// SettingDef describes one company setting: its JSON type, default value and
// (for enums) the allowed values. The catalog is the single source of truth;
// the front keeps a copy of the defaults in constants/company-settings.js.
type SettingDef struct {
	Type    string // "bool", "int", "enum"
	Default interface{}
	Values  []string
	Min     *int
}

var ErrUnknownSetting = errors.New("unknown setting")

var zero = 0

var SettingsCatalog = map[string]SettingDef{
	"permissions_mode":        {Type: "enum", Default: "individual", Values: []string{"role", "individual"}},
	"order_card_code":         {Type: "bool", Default: false},
	"lock_delivered_status":   {Type: "bool", Default: false},
	"deliver_without_payment": {Type: "bool", Default: false},
	"payment_tolerance":       {Type: "int", Default: 0, Min: &zero},
	"volume_mode":             {Type: "enum", Default: "type", Values: []string{"type", "worker"}},
	"volume_skip_countable":   {Type: "bool", Default: false},
	"kpi_enabled":             {Type: "bool", Default: false},
	"storage_slots":           {Type: "bool", Default: false},
}

// DefaultSettings returns every catalog key with its default as JSON.
func DefaultSettings() map[string]json.RawMessage {
	out := make(map[string]json.RawMessage, len(SettingsCatalog))
	for key, def := range SettingsCatalog {
		b, _ := json.Marshal(def.Default)
		out[key] = b
	}
	return out
}

// ValidateSetting checks that raw is a valid JSON value for key and returns it
// normalised (whitespace trimmed, numbers re-encoded).
func ValidateSetting(key string, raw json.RawMessage) (json.RawMessage, error) {
	def, ok := SettingsCatalog[key]
	if !ok {
		return nil, ErrUnknownSetting
	}
	raw = bytes.TrimSpace(raw)
	switch def.Type {
	case "bool":
		var v bool
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, fmt.Errorf("%s must be true or false", key)
		}
		return json.Marshal(v)
	case "int":
		var any interface{}
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()
		if err := dec.Decode(&any); err != nil {
			return nil, fmt.Errorf("%s must be an integer", key)
		}
		n, ok := any.(json.Number)
		if !ok {
			return nil, fmt.Errorf("%s must be an integer", key)
		}
		v, err := n.Int64()
		if err != nil {
			return nil, fmt.Errorf("%s must be an integer", key)
		}
		if def.Min != nil && v < int64(*def.Min) {
			return nil, fmt.Errorf("%s must be >= %d", key, *def.Min)
		}
		return json.Marshal(v)
	case "enum":
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, fmt.Errorf("%s must be one of %v", key, def.Values)
		}
		for _, allowed := range def.Values {
			if v == allowed {
				return json.Marshal(v)
			}
		}
		return nil, fmt.Errorf("%s must be one of %v", key, def.Values)
	}
	return nil, ErrUnknownSetting
}

// MergeSettings lays the stored rows over the defaults so every key is present.
func MergeSettings(rows map[string]json.RawMessage) map[string]json.RawMessage {
	out := DefaultSettings()
	for key, value := range rows {
		if _, ok := SettingsCatalog[key]; ok {
			out[key] = value
		}
	}
	return out
}

type UpdateCompanySettingRequest struct {
	Key   string          `json:"key" binding:"required"`
	Value json.RawMessage `json:"value" binding:"required"`
}
