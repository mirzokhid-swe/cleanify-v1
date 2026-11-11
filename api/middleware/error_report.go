package middleware

import (
	"bw-erp/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Har qanday xatolik (javob kodi >= 400 yoki panic) Telegram guruhga yuboriladi:
// qaysi API, qaysi handler, qaysi korxona va foydalanuvchi, so'rov tanasi, javob.
// Yo'l topilmagan 404 (bot skanerlari) hisobot qilinmaydi. Yuborish asinxron,
// so'rovni sekinlashtirmaydi; Telegram ishlamasa faqat log.

const (
	errReportMaxBody    = 64 << 10 // shundan katta so'rov tanasi (fayl, ovoz) o'qilmaydi
	errReportMaxCapture = 2 << 10  // javobdan shuncha bayt saqlanadi
	errReportMaxText    = 3900     // Telegram chegarasi 4096
)

type ErrorInfo struct {
	Env       string
	At        time.Time
	Method    string
	URL       string
	Handler   string
	Status    int
	Duration  time.Duration
	User      *models.User
	CompanyID string
	Response  string
	Body      string
	IP        string
	UserAgent string
	Stack     string
}

func ErrorReport(env string, send func(string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength <= errReportMaxBody {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		w := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = w
		start := time.Now()
		defer func() {
			r := recover()
			if r == nil && (c.Writer.Status() < 400 || c.FullPath() == "") {
				return
			}
			info := collectErrorInfo(c, env, body, w.buf.String(), start)
			if r != nil {
				info.Status = 500
				info.Response = fmt.Sprint("panic: ", r)
				info.Stack = string(debug.Stack())
			}
			send(FormatErrorReport(info))
			if r != nil {
				panic(r)
			}
		}()
		c.Next()
	}
}

type captureWriter struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *captureWriter) Write(b []byte) (int, error) {
	if w.buf.Len() < errReportMaxCapture {
		w.buf.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *captureWriter) WriteString(s string) (int, error) {
	if w.buf.Len() < errReportMaxCapture {
		w.buf.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

func collectErrorInfo(c *gin.Context, env string, body []byte, response string, start time.Time) ErrorInfo {
	info := ErrorInfo{
		Env:       env,
		At:        start,
		Method:    c.Request.Method,
		URL:       c.Request.URL.RequestURI(),
		Handler:   strings.TrimPrefix(c.HandlerName(), "bw-erp/api/"),
		Status:    c.Writer.Status(),
		Duration:  time.Since(start),
		Response:  response,
		Body:      string(body),
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	if v, ok := c.Get("user"); ok {
		if u, ok := v.(models.User); ok {
			info.User = &u
		}
	}
	if v, ok := c.Get("company_id"); ok {
		info.CompanyID, _ = v.(string)
	}
	if info.CompanyID == "" {
		info.CompanyID = c.Query("company_id")
	}
	if info.CompanyID == "" && len(body) > 0 {
		var b struct {
			CompanyID string `json:"company_id"`
		}
		if json.Unmarshal(body, &b) == nil {
			info.CompanyID = b.CompanyID
		}
	}
	return info
}

var passwordRe = regexp.MustCompile(`(?i)("[^"]*password[^"]*"\s*:\s*)"[^"]*"`)

func maskPasswords(s string) string {
	return passwordRe.ReplaceAllString(s, `$1"***"`)
}

func FormatErrorReport(i ErrorInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "❗ Go API xatolik [%s]\n", i.Env)
	fmt.Fprintf(&b, "Vaqt: %s\n", i.At.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "So'rov: %s %s → %d (%dms)\n", i.Method, i.URL, i.Status, i.Duration.Milliseconds())
	if i.Handler != "" {
		fmt.Fprintf(&b, "Handler: %s\n", i.Handler)
	}
	company := i.CompanyID
	if i.User != nil {
		for _, uc := range i.User.UserPermissionByCompany {
			if uc.CompanyID == i.CompanyID && uc.CompanyName != "" {
				company = uc.CompanyName + " (" + i.CompanyID + ")"
			}
		}
	}
	if company != "" {
		fmt.Fprintf(&b, "Korxona: %s\n", company)
	}
	if i.User != nil {
		fmt.Fprintf(&b, "Foydalanuvchi: #%d %s, %s", i.User.ID, i.User.Fullname, i.User.Phone)
		if i.User.IsSuperAdmin {
			b.WriteString(" (super admin)")
		}
		b.WriteString("\n")
	} else {
		b.WriteString("Foydalanuvchi: kirmagan (token yo'q yoki noto'g'ri)\n")
	}
	if i.Response != "" {
		fmt.Fprintf(&b, "Javob: %s\n", strings.TrimSpace(i.Response))
	}
	if i.Body != "" {
		fmt.Fprintf(&b, "So'rov tanasi: %s\n", maskPasswords(strings.TrimSpace(i.Body)))
	}
	fmt.Fprintf(&b, "IP: %s\nUA: %s\n", i.IP, i.UserAgent)
	if i.Stack != "" {
		fmt.Fprintf(&b, "Stack:\n%s", i.Stack)
	}
	text := b.String()
	if len(text) > errReportMaxText {
		text = text[:errReportMaxText] + "…"
	}
	return text
}

var telegramAPI = "https://api.telegram.org"

// TelegramSender: xabarni guruhga goroutine'da yuboradi. Token yoki chat_id
// bo'lmasa faqat log'ga yozadi (lokal muhit).
func TelegramSender(token, chatID string) func(string) {
	if token == "" || chatID == "" {
		return func(text string) {
			log.Println("[error-report] telegram o'chiq:", strings.SplitN(text, "\n", 4)[2])
		}
	}
	client := &http.Client{Timeout: 5 * time.Second}
	return func(text string) {
		go func() {
			resp, err := client.PostForm(telegramAPI+"/bot"+token+"/sendMessage", url.Values{
				"chat_id":                  {chatID},
				"text":                     {text},
				"disable_web_page_preview": {"true"},
			})
			if err != nil {
				log.Println("[error-report] telegram:", err)
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Println("[error-report] telegram status:", resp.StatusCode)
			}
		}()
	}
}
