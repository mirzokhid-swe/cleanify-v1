package handlers

import (
	"bw-erp/api/http"
	"bw-erp/api/middleware"
	"bw-erp/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Order item handlers learn the company from the order, not from the request,
// so each one loads the order and checks the permission with UserCan.

func (h *Handler) CreateOrderItemModel(c *gin.Context) {
	var body models.CreateOrderItemModel
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	order, err := h.Stg.Order().GetByPrimaryKey(body.OrderID)
	if err != nil {
		h.handleResponse(c, http.NOT_FOUND, "Order not found")
		return
	}

	if !middleware.UserCan(c, order.CompanyID, "create_order_item") {
		h.handleResponse(c, http.PermissionDenied, "Permission denied")
		return
	}

	// The type must belong to the order's company.
	orderItemType, err := h.Stg.OrderItemType().GetById(body.OrderItemTypeID, order.CompanyID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, "order item type not found")
		return
	}
	body.ItemType = orderItemType.Name
	body.IsCountable = orderItemType.IsCountable

	if !body.IsCountable && (body.Width == 0 && body.Height == 0) {
		h.handleResponse(c, http.BadRequest, "width and height is required")
		return
	}

	user := c.MustGet("user").(models.User)
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

	orderID, err := h.Stg.OrderItem().GetOrderID(body.ID)
	if err != nil {
		h.handleResponse(c, http.NOT_FOUND, "Order item not found")
		return
	}
	order, err := h.Stg.Order().GetByPrimaryKey(orderID)
	if err != nil {
		h.handleResponse(c, http.NOT_FOUND, "Order not found")
		return
	}

	if !middleware.UserCan(c, order.CompanyID, "edit_order_item") {
		h.handleResponse(c, http.PermissionDenied, "Permission denied")
		return
	}

	orderItemType, err := h.Stg.OrderItemType().GetById(body.OrderItemTypeID, order.CompanyID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, "order item type not found")
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
	orderItemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	orderID, err := h.Stg.OrderItem().GetOrderID(orderItemID)
	if err != nil {
		h.handleResponse(c, http.NOT_FOUND, "Order item not found")
		return
	}
	order, err := h.Stg.Order().GetByPrimaryKey(orderID)
	if err != nil {
		h.handleResponse(c, http.NOT_FOUND, "Order not found")
		return
	}

	if !middleware.UserCan(c, order.CompanyID, "delete_order_item") {
		h.handleResponse(c, http.PermissionDenied, "Permission denied")
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

	orderID, err := h.Stg.OrderItem().GetOrderID(body.ID)
	if err != nil {
		h.handleResponse(c, http.NOT_FOUND, "Order item not found")
		return
	}
	order, err := h.Stg.Order().GetByPrimaryKey(orderID)
	if err != nil {
		h.handleResponse(c, http.NOT_FOUND, "Order not found")
		return
	}

	if !middleware.UserCan(c, order.CompanyID, "edit_order_item") {
		h.handleResponse(c, http.PermissionDenied, "Permission denied")
		return
	}

	user := c.MustGet("user").(models.User)
	_, err = h.Stg.OrderItem().UpdateStatus(user.ID, body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, "Update successfully!")
}
