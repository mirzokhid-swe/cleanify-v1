package handlers

import (
	"bw-erp/api/http"
	"bw-erp/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPermissionList(c *gin.Context) {
	perms, err := h.Stg.Permission().GetList("company")
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	// company_id is optional: with it, status permissions get the company's own status names.
	var statuses []models.OrderStatus
	if companyID := c.Query("company_id"); companyID != "" {
		statuses, err = h.Stg.OrderStatus().GetList(companyID, "number", "asc")
		if err != nil {
			h.handleResponse(c, http.BadRequest, err.Error())
			return
		}
	}

	h.handleResponse(c, http.OK, models.GroupPermissions(perms, statuses))
}
