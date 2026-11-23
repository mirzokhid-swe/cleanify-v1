package handlers

import (
	"bw-erp/api/http"
	"bw-erp/config"
	"bw-erp/storage"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Stg storage.StorageI
	Cfg config.Config
	// Front xatoliklari shu funksiya bilan yuboriladi (SetUpRouter Telegram sender qo'yadi).
	ClientErrorSend func(string)
}

func NewHandler(stg storage.StorageI, cfg config.Config) Handler {
	return Handler{
		Stg: stg,
		Cfg: cfg,
	}
}

func (h *Handler) handleResponse(c *gin.Context, status http.Status, data interface{}) {
	c.JSON(status.Code, http.Response{
		Status:      status.Status,
		Description: status.Description,
		Data:        data,
	})
}

func (h *Handler) getStatusParam(c *gin.Context) (offset int, err error) {
	offsetStr := c.DefaultQuery("status", "0")
	return strconv.Atoi(offsetStr)
}

func (h *Handler) getOffsetParam(c *gin.Context) (offset int, err error) {
	offsetStr := c.DefaultQuery("offset", h.Cfg.DefaultOffset)
	return strconv.Atoi(offsetStr)
}

func (h *Handler) getLimitParam(c *gin.Context) (offset int, err error) {
	offsetStr := c.DefaultQuery("limit", h.Cfg.DefaultLimit)
	return strconv.Atoi(offsetStr)
}
