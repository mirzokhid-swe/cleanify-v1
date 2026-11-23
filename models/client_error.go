package models

import (
	"fmt"
	"strings"
	"time"
)

// Front (brauzer) o'zida yuz bergan xatolik: JS xatosi, ushlanmagan promise,
// React render xatosi. Front buni POST /api/v1/client-errors ga yuboradi,
// Go uni xatoliklar Telegram guruhiga jo'natadi.

const (
	ClientErrorMaxBody        = 32 << 10 // so'rov tanasi chegarasi
	ClientErrorMaxPerMinute   = 20       // Telegramga bir daqiqada shuncha xabar
	ClientErrorMaxText        = 3900     // Telegram chegarasi 4096
	ClientErrorMaxBreadcrumbs = 15
)

type ClientErrorBreadcrumb struct {
	At   string `json:"at"`   // qurilma vaqti, "15:04:05"
	Type string `json:"type"` // nav, click, api
	Text string `json:"text"`
}

type ClientError struct {
	Type           string                  `json:"type"` // error, unhandledrejection, react
	Message        string                  `json:"message" binding:"required"`
	Stack          string                  `json:"stack"`
	ComponentStack string                  `json:"component_stack"`
	URL            string                  `json:"url"`
	Build          string                  `json:"build"`
	At             string                  `json:"at"`    // birinchi marta yuz bergan vaqt (qurilma soati)
	Count          int                     `json:"count"` // yuborilguncha necha marta takrorlandi
	UserID         int                     `json:"user_id"`
	UserName       string                  `json:"user_name"`
	UserPhone      string                  `json:"user_phone"`
	CompanyID      string                  `json:"company_id"`
	CompanyName    string                  `json:"company_name"`
	Screen         string                  `json:"screen"`
	Online         bool                    `json:"online"`
	Lang           string                  `json:"lang"`
	Telegram       string                  `json:"telegram"`
	Breadcrumbs    []ClientErrorBreadcrumb `json:"breadcrumbs"`
}

func clientErrorCut(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

// FormatClientError: Telegram xabari matni. Foydalanuvchi va korxona frontning
// o'zi aytgani (token tekshirilmaydi: xatolik kirishdan oldin ham bo'lishi mumkin).
func FormatClientError(env string, now time.Time, e ClientError, ip string, userAgent string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "🖥 Front xatolik [%s]\n", env)
	fmt.Fprintf(&b, "Vaqt: %s", now.In(time.FixedZone("Asia/Tashkent", 5*3600)).Format("2006-01-02 15:04:05"))
	if e.At != "" {
		fmt.Fprintf(&b, " (qurilmada: %s)", clientErrorCut(e.At, 40))
	}
	b.WriteString("\n")
	if e.Count > 1 {
		fmt.Fprintf(&b, "Takror: %d marta\n", e.Count)
	}
	if e.Type != "" {
		fmt.Fprintf(&b, "Tur: %s\n", clientErrorCut(e.Type, 40))
	}
	fmt.Fprintf(&b, "Xabar: %s\n", clientErrorCut(e.Message, 300))
	if e.URL != "" {
		fmt.Fprintf(&b, "Sahifa: %s\n", clientErrorCut(e.URL, 200))
	}
	if e.Build != "" {
		fmt.Fprintf(&b, "Build: %s\n", clientErrorCut(e.Build, 40))
	}
	if e.CompanyID != "" || e.CompanyName != "" {
		fmt.Fprintf(&b, "Korxona: %s (%s)\n", clientErrorCut(e.CompanyName, 80), clientErrorCut(e.CompanyID, 40))
	}
	if e.UserID != 0 {
		fmt.Fprintf(&b, "Foydalanuvchi: #%d %s, %s\n", e.UserID, clientErrorCut(e.UserName, 80), clientErrorCut(e.UserPhone, 20))
	} else {
		b.WriteString("Foydalanuvchi: kirmagan\n")
	}
	online := "online"
	if !e.Online {
		online = "offline"
	}
	fmt.Fprintf(&b, "Qurilma: %s, %s, %s", clientErrorCut(e.Screen, 20), online, clientErrorCut(e.Lang, 20))
	if e.Telegram != "" {
		fmt.Fprintf(&b, ", Telegram %s", clientErrorCut(e.Telegram, 40))
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "IP: %s\nUA: %s\n", ip, clientErrorCut(userAgent, 200))
	crumbs := e.Breadcrumbs
	if len(crumbs) > ClientErrorMaxBreadcrumbs {
		crumbs = crumbs[len(crumbs)-ClientErrorMaxBreadcrumbs:]
	}
	if len(crumbs) > 0 {
		b.WriteString("Oxirgi harakatlar:\n")
		for _, c := range crumbs {
			fmt.Fprintf(&b, "%s %s %s\n", clientErrorCut(c.At, 8), clientErrorCut(c.Type, 8), clientErrorCut(c.Text, 110))
		}
	}
	if e.Stack != "" {
		fmt.Fprintf(&b, "Stack:\n%s\n", clientErrorCut(e.Stack, 1200))
	}
	if e.ComponentStack != "" {
		fmt.Fprintf(&b, "Komponentlar:\n%s\n", clientErrorCut(e.ComponentStack, 500))
	}
	return clientErrorCut(b.String(), ClientErrorMaxText)
}
