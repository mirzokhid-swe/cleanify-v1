package handlers

import (
	"database/sql"
	"errors"

	"bw-erp/api/http"
	"bw-erp/models"
	"bw-erp/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) CreateOrderItemTypeModel(c *gin.Context) {
	var body models.OrderItemTypeModel

	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if !utils.IsValidUUID(body.CompanyID) {
		h.handleResponse(c, http.InvalidArgument, "company id is an invalid uuid")
		return
	}

	_, err := h.Stg.Company().GetById(body.CompanyID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	id := uuid.New()

	err = h.Stg.OrderItemType().Create(id.String(), body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.Created, id)
}

func (h *Handler) GetOrderItemTypesByCompany(c *gin.Context) {
	var body models.GetOrderItemTypeByCompany
	if err := c.ShouldBindQuery(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	data, err := h.Stg.OrderItemType().GetByCompany(body.CompanyID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, data)
}

func (h *Handler) GetOrderItemTypeByID(c *gin.Context) {
	ID := c.Param("id")
	if !utils.IsValidUUID(ID) {
		h.handleResponse(c, http.InvalidArgument, "id is an invalid uuid")
		return
	}

	data, err := h.Stg.OrderItemType().GetById(ID, c.Query("company_id"))
	if errors.Is(err, sql.ErrNoRows) {
		h.handleResponse(c, http.NOT_FOUND, "Order item type not found!")
		return
	}
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, data)
}

func (h *Handler) UpdateOrderItemType(c *gin.Context) {
	var body models.EditOrderItemTypeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	rowsAffected, err := h.Stg.OrderItemType().Update(body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if rowsAffected == 0 {
		h.handleResponse(c, http.NOT_FOUND, "Order item type not found!")
		return
	}

	h.handleResponse(c, http.OK, rowsAffected)
}

func (h *Handler) DeleteOrderItemType(c *gin.Context) {
	ID := c.Param("id")
	if !utils.IsValidUUID(ID) {
		h.handleResponse(c, http.InvalidArgument, "id is an invalid uuid")
		return
	}

	var body models.DeleteOrderItemTypeRequest
	if err := c.ShouldBindQuery(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	rowsAffected, err := h.Stg.OrderItemType().Delete(ID, body.CompanyID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if rowsAffected == 0 {
		h.handleResponse(c, http.NOT_FOUND, "Order item type not found!")
		return
	}

	h.handleResponse(c, http.OK, rowsAffected)
}
