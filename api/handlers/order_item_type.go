package handlers

import (
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
	var body models.GetOrderStatusListRequest
	if err := c.ShouldBindQuery(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	// token, err := utils.ExtractTokenID(c)

	// if err != nil {
	// 	h.handleResponse(c, http.BadRequest, err.Error())
	// 	return
	// }

	// user, err := h.Stg.User().GetById(token.UserID)
	// if err != nil {
	// 	h.handleResponse(c, http.BadRequest, err.Error())
	// 	return
	// }

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
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, data)
}

func (h *Handler) UpdateOrderItemType(c *gin.Context) {
	// token, err := utils.ExtractTokenID(c)
	var body models.EditOrderItemTypeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	// if err != nil {
	// 	h.handleResponse(c, http.BadRequest, err.Error())
	// 	return
	// }

	// user, err := h.Stg.User().GetById(token.UserID)
	// if err != nil {
	// 	h.handleResponse(c, http.BadRequest, err.Error())
	// 	return
	// }
	// body.CopmanyID = *user.CompanyID

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
