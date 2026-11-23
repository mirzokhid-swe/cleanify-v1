package handlers

import (
	"bw-erp/api/http"
	"bw-erp/models"
	"bw-erp/pkg/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Kod 6 xonali va tokensiz tekshiriladi, shuning uchun qo'pol kuchga qarshi
// har IP'dan bir daqiqada ko'pi bilan authCodeMaxPerMinute urinish.
const authCodeMaxPerMinute = 10

var (
	authCodeMu     sync.Mutex
	authCodeMinute int64
	authCodeCounts = map[string]int{}
)

func authCodeAllowed(ip string, now time.Time) bool {
	authCodeMu.Lock()
	defer authCodeMu.Unlock()
	minute := now.Unix() / 60
	if minute != authCodeMinute {
		authCodeMinute = minute
		authCodeCounts = map[string]int{}
	}
	authCodeCounts[ip]++
	return authCodeCounts[ip] <= authCodeMaxPerMinute
}

// ResetAuthCodeLimiter: faqat testlar uchun.
func ResetAuthCodeLimiter() {
	authCodeMu.Lock()
	defer authCodeMu.Unlock()
	authCodeMinute = 0
	authCodeCounts = map[string]int{}
}

// VerifyAuthCode: auth bot bergan bir martalik kodni JWT'ga almashtiradi.
// Token talab qilinmaydi — bu kirishning o'zi.
func (h *Handler) VerifyAuthCode(c *gin.Context) {
	if !authCodeAllowed(c.ClientIP(), time.Now()) {
		h.handleResponse(c, http.TooManyRequests, "Urinishlar tugadi")
		return
	}

	var body models.VerifyAuthCodeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, "Kod noto'g'ri yoki eskirgan")
		return
	}

	userID, err := h.Stg.AuthCode().Use(body.Code)
	if err != nil {
		h.handleResponse(c, http.BadRequest, "Kod noto'g'ri yoki eskirgan")
		return
	}

	user, err := h.Stg.User().GetById(userID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, "Kod noto'g'ri yoki eskirgan")
		return
	}

	accessToken, refreshToken, err := utils.GenerateToken(strconv.FormatInt(user.ID, 10), user.Phone)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, models.AuthorizationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
