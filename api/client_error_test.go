package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"bw-erp/api/handlers"
	"bw-erp/config"
)

func clientErrorRouter(sent *[]string) *handlers.Handler {
	h := handlers.Handler{}
	h.ClientErrorSend = func(s string) { *sent = append(*sent, s) }
	return &h
}

func TestReportClientErrorSendsToTelegram(t *testing.T) {
	var sent []string
	h := clientErrorRouter(&sent)
	r := SetUpRouter(*h, config.Config{Release_Mode: "test"})
	body := `{"type":"error","message":"TypeError: x is undefined","stack":"at a.js:1:2","url":"/orders","user_id":7,"user_name":"Azizbek","company_id":"c1","company_name":"Yangi Dunyo","breadcrumbs":[{"at":"14:00:01","type":"click","text":"button \"Saqlash\""}]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/client-errors", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "cleanify-app")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("200 kutildi, %d %s", w.Code, w.Body.String())
	}
	if len(sent) != 1 {
		t.Fatalf("1 ta xabar kutildi, %d", len(sent))
	}
	for _, want := range []string{"Front xatolik", "TypeError: x is undefined", "Sahifa: /orders", "#7 Azizbek", "Yangi Dunyo (c1)", "UA: cleanify-app", "14:00:01 click button \"Saqlash\"", "at a.js:1:2"} {
		if !strings.Contains(sent[0], want) {
			t.Fatalf("xabarda %q yo'q:\n%s", want, sent[0])
		}
	}
}

func TestReportClientErrorRejectsBadBody(t *testing.T) {
	for name, body := range map[string]string{
		"xabarsiz":   `{"type":"error"}`,
		"json emas":  `salom`,
		"juda katta": `{"message":"x","stack":"` + strings.Repeat("s", 40<<10) + `"}`,
	} {
		var sent []string
		h := clientErrorRouter(&sent)
		r := SetUpRouter(*h, config.Config{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/client-errors", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("%s: 400 kutildi, %d", name, w.Code)
		}
		for _, s := range sent {
			if strings.Contains(s, "Front xatolik") {
				t.Fatalf("%s: front xabari yuborilmasligi kerak", name)
			}
		}
	}
}
