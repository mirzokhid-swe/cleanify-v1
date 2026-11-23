package handlers

import (
	"bw-erp/api/http"
	"bw-erp/models"
	"log"
	nethttp "net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Telegramga bir daqiqada ClientErrorMaxPerMinute tadan ko'p front xatoligi
// yuborilmaydi (bitta qurilmadagi sikl guruhni to'ldirib yubormasin).
var (
	clientErrorMu     sync.Mutex
	clientErrorMinute int64
	clientErrorCount  int
)

func clientErrorAllowed(now time.Time) bool {
	clientErrorMu.Lock()
	defer clientErrorMu.Unlock()
	minute := now.Unix() / 60
	if minute != clientErrorMinute {
		clientErrorMinute = minute
		clientErrorCount = 0
	}
	clientErrorCount++
	return clientErrorCount <= models.ClientErrorMaxPerMinute
}

// ReportClientError: front o'zida yuz bergan xatolikni yuboradi. Token talab
// qilinmaydi — xatolik kirishdan oldin ham bo'lishi mumkin. Chegaradan oshgan
// xabar ham 200 oladi (front qayta yubormasin), faqat log'ga yoziladi.
func (h *Handler) ReportClientError(c *gin.Context) {
	c.Request.Body = nethttp.MaxBytesReader(c.Writer, c.Request.Body, models.ClientErrorMaxBody)
	var e models.ClientError
	if err := c.ShouldBindJSON(&e); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	now := time.Now()
	text := models.FormatClientError(h.Cfg.Release_Mode, now, e, c.ClientIP(), c.Request.UserAgent())
	if h.ClientErrorSend == nil || !clientErrorAllowed(now) {
		log.Println("[client-error] yuborilmadi:", e.Message)
		h.handleResponse(c, http.OK, "ok")
		return
	}
	h.ClientErrorSend(text)
	h.handleResponse(c, http.OK, "ok")
}
