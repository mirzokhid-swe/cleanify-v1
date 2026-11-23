package models

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestFormatClientError(t *testing.T) {
	now := time.Date(2026, 11, 22, 9, 3, 5, 0, time.UTC)
	e := ClientError{
		Type:           "react",
		Message:        "TypeError: Cannot read properties of undefined (reading 'map')",
		Stack:          "TypeError: x\n    at OrderList (https://app/assets/index-abc.js:40:1234)",
		ComponentStack: "\n    at OrderList\n    at MainLayout",
		URL:            "/orders/15?tab=2",
		Build:          "2026-11-21T10:00:00.000Z",
		At:             "2026-11-22T14:03:01+05:00",
		Count:          3,
		UserID:         7,
		UserName:       "Azizbek",
		UserPhone:      "+998901112233",
		CompanyID:      "c1",
		CompanyName:    "Yangi Dunyo",
		Screen:         "390x844",
		Online:         true,
		Lang:           "uz",
		Telegram:       "android 7.2",
		Breadcrumbs: []ClientErrorBreadcrumb{
			{At: "14:02:58", Type: "nav", Text: "/orders"},
			{At: "14:03:00", Type: "click", Text: "button \"Saqlash\""},
			{At: "14:03:01", Type: "api", Text: "POST orders/edit → 200"},
		},
	}
	got := FormatClientError("prod", now, e, "10.0.0.1", "cleanify-app")
	for _, want := range []string{
		"🖥 Front xatolik [prod]",
		"Vaqt: 2026-11-22 14:03:05 (qurilmada: 2026-11-22T14:03:01+05:00)",
		"Takror: 3 marta",
		"Tur: react",
		"Xabar: TypeError: Cannot read properties of undefined (reading 'map')",
		"Sahifa: /orders/15?tab=2",
		"Build: 2026-11-21T10:00:00.000Z",
		"Korxona: Yangi Dunyo (c1)",
		"Foydalanuvchi: #7 Azizbek, +998901112233",
		"Qurilma: 390x844, online, uz, Telegram android 7.2",
		"IP: 10.0.0.1\nUA: cleanify-app",
		"Oxirgi harakatlar:\n14:02:58 nav /orders\n14:03:00 click button \"Saqlash\"\n14:03:01 api POST orders/edit → 200",
		"Stack:\nTypeError: x\n    at OrderList (https://app/assets/index-abc.js:40:1234)",
		"Komponentlar:\nat OrderList\n    at MainLayout",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("xabarda %q yo'q:\n%s", want, got)
		}
	}
}

func TestFormatClientErrorAnonymousAndMinimal(t *testing.T) {
	got := FormatClientError("dev", time.Now(), ClientError{Message: "boom"}, "ip", "ua")
	for _, want := range []string{"Xabar: boom", "Foydalanuvchi: kirmagan", "offline"} {
		if !strings.Contains(got, want) {
			t.Fatalf("xabarda %q yo'q:\n%s", want, got)
		}
	}
	for _, not := range []string{"Takror", "Korxona", "Stack", "Komponentlar", "Oxirgi harakatlar", "Sahifa", "Build", "qurilmada"} {
		if strings.Contains(got, not) {
			t.Fatalf("bo'sh maydon chiqmasligi kerak (%q):\n%s", not, got)
		}
	}
}

func TestFormatClientErrorLimits(t *testing.T) {
	e := ClientError{
		Message:        strings.Repeat("ў", 5000),
		Stack:          strings.Repeat("s", 50000),
		ComponentStack: strings.Repeat("c", 50000),
	}
	for i := 0; i < 40; i++ {
		e.Breadcrumbs = append(e.Breadcrumbs, ClientErrorBreadcrumb{At: "14:00:00", Type: "click", Text: strings.Repeat("b", 500)})
	}
	e.Breadcrumbs[39].Text = "oxirgisi"
	got := FormatClientError("prod", time.Now(), e, "ip", "ua")
	if n := utf8.RuneCountInString(got); n > ClientErrorMaxText+1 {
		t.Fatalf("xabar juda uzun: %d", n)
	}
	if !utf8.ValidString(got) {
		t.Fatal("kesilganda UTF-8 buzildi")
	}
	if strings.Count(got, "14:00:00 click") != ClientErrorMaxBreadcrumbs || !strings.Contains(got, "oxirgisi") {
		t.Fatalf("oxirgi %d ta harakat qolishi kerak:\n%s", ClientErrorMaxBreadcrumbs, got)
	}
	if !strings.Contains(got, "Stack:\nsss") {
		t.Fatalf("stack sig'ishi kerak")
	}
}
