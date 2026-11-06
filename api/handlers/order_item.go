package handlers

import (
	"bw-erp/api/http"
	"bw-erp/models"
	"bw-erp/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateOrderItemModel(c *gin.Context) {
	var body models.CreateOrderItemModel
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	_, err := h.Stg.Order().GetByPrimaryKey(body.OrderID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, "Order not found")
		return
	}

	orderItemType, err := h.Stg.OrderItemType().GetById(body.OrderItemTypeID, "")
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	body.ItemType = orderItemType.Name
	body.IsCountable = orderItemType.IsCountable

	if !body.IsCountable && (body.Width == 0 && body.Height == 0) {
		h.handleResponse(c, http.BadRequest, "width and height is required")
		return
	}

	token, err := utils.ExtractTokenID(c)

	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	user, err := h.Stg.User().GetById(token.UserID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	err = h.Stg.OrderItem().Create(user.ID, body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.Created, "Created successfully!")
}

func (h *Handler) UpdateOrderItemModel(c *gin.Context) {
	var body models.UpdateOrderItemRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	orderItemType, err := h.Stg.OrderItemType().GetById(body.OrderItemTypeID, "")
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	body.ItemType = orderItemType.Name
	body.IsCountable = orderItemType.IsCountable

	_, err = h.Stg.OrderItem().Update(body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, "Update successfully!")
}

func (h *Handler) DeleteOrderItemByID(c *gin.Context) {
	ID := c.Param("id")
	orderItemID, err := strconv.Atoi(ID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	err = h.Stg.OrderItem().DeleteByID(orderItemID)

	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, "Deleted successfully!")
}

func (h *Handler) UpdateOrderItemStatus(c *gin.Context) {
	var body models.UpdateOrderItemStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	token, err := utils.ExtractTokenID(c)

	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	user, err := h.Stg.User().GetById(token.UserID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	_, err = h.Stg.OrderItem().UpdateStatus(user.ID, body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, "Update successfully!")
}
