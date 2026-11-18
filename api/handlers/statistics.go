package handlers

import (
	"bw-erp/api/http"
	"bw-erp/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetWorkVolumeList(c *gin.Context) {
	var body models.GetWorkVolumeListRequest
	if err := c.ShouldBindQuery(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	data, err := h.Stg.Statistics().GetWorkVolume(body.CompanyID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, data)
}

func (h *Handler) GetStatusMetrics(c *gin.Context) {
	var body models.StatusMetricsRequest
	if err := c.ShouldBindQuery(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	days := 40
	if company, err := h.Stg.Company().GetShow(body.CompanyID); err == nil {
		days = models.StatusMetricsDays(company.Settings)
	}

	data, err := h.Stg.Statistics().GetStatusMetrics(body.CompanyID, body.Status, days)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, data)
}

func (h *Handler) GetServicePaymentStatistics(c *gin.Context) {
	var body models.GetServicePaymentStatisticsRequest
	if err := c.ShouldBindQuery(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	data, err := h.Stg.Statistics().GetServicePaymentStatistics(body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, data)
}
