package middleware

import (
	"bw-erp/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func errReportRouter(sent *[]string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery(), ErrorReport("test", func(s string) { *sent = append(*sent, s) }))
	r.POST("/api/v1/orders/edit", func(c *gin.Context) {
		c.Set("user", models.User{ID: 7, Fullname: "Azizbek", Phone: "+998901112233", UserPermissionByCompany: []models.UserPermissionByCompany{{CompanyID: "c1", CompanyName: "Yangi Dunyo"}}})
		c.Set("company_id", "c1")
		c.JSON(400, gin.H{"status": "BAD_REQUEST", "data": "joy raqami noto'g'ri"})
	})
	r.GET("/ok", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.GET("/boom", func(c *gin.Context) { panic("kutilmagan") })
	r.POST("/login", func(c *gin.Context) { c.JSON(401, gin.H{"message": "wrong"}) })
	return r
}

func TestErrorReportSendsContext(t *testing.T) {
	var sent []string
	r := errReportRouter(&sent)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/orders/edit?x=1", strings.NewReader(`{"id":5,"slots":["12"]}`))
	req.Header.Set("User-Agent", "cleanify-app")
	r.ServeHTTP(w, req)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "joy raqami") {
		t.Fatalf("javob buzildi: %d %s", w.Code, w.Body.String())
	}
	if len(sent) != 1 {
		t.Fatalf("1 ta xabar kutildi, %d", len(sent))
	}
	for _, want := range []string{"[test]", "POST /api/v1/orders/edit?x=1 → 400", "Korxona: Yangi Dunyo (c1)", "#7 Azizbek, +998901112233", `"slots":["12"]`, "joy raqami noto'g'ri", "UA: cleanify-app", "Handler: middleware."} {
		if !strings.Contains(sent[0], want) {
			t.Fatalf("xabarda %q yo'q:\n%s", want, sent[0])
		}
	}
}

func TestErrorReportSkipsOkAndUnknownRoute(t *testing.T) {
	var sent []string
	r := errReportRouter(&sent)
	for _, p := range []string{"/ok", "/wp-admin"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
	}
	if len(sent) != 0 {
		t.Fatalf("200 va yo'l topilmagan 404 hisobot qilinmasligi kerak: %v", sent)
	}
}

func TestErrorReportPanic(t *testing.T) {
	var sent []string
	r := errReportRouter(&sent)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/boom", nil))
	if w.Code != 500 {
		t.Fatalf("Recovery 500 berishi kerak, %d", w.Code)
	}
	if len(sent) != 1 || !strings.Contains(sent[0], "→ 500") || !strings.Contains(sent[0], "panic: kutilmagan") || !strings.Contains(sent[0], "Stack:") {
		t.Fatalf("panic hisoboti noto'g'ri: %v", sent)
	}
}

func TestErrorReportMasksPasswordAndAnonymous(t *testing.T) {
	var sent []string
	r := errReportRouter(&sent)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/login", strings.NewReader(`{"phone":"+998","password":"secret1","new_password": "s2"}`)))
	if len(sent) != 1 || strings.Contains(sent[0], "secret1") || strings.Contains(sent[0], "s2\"") || !strings.Contains(sent[0], `"password":"***"`) || !strings.Contains(sent[0], "kirmagan") {
		t.Fatalf("parol yashirilmadi: %v", sent)
	}
}

func TestTelegramSender(t *testing.T) {
	var mu sync.Mutex
	got := url_Values{}
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		mu.Lock()
		got = url_Values{"path": r.URL.Path, "chat_id": r.Form.Get("chat_id"), "text": r.Form.Get("text")}
		mu.Unlock()
		close(done)
	}))
	defer srv.Close()
	telegramAPI = srv.URL
	TelegramSender("TOKEN", "-100")("salom")
	<-done
	mu.Lock()
	defer mu.Unlock()
	if got["path"] != "/botTOKEN/sendMessage" || got["chat_id"] != "-100" || got["text"] != "salom" {
		t.Fatal(got)
	}
}

type url_Values map[string]string
