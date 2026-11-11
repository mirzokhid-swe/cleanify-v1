package repo

import "bw-erp/models"

type OrderI interface {
	Create(userID int64, entity models.CreateOrderModel) (id int, err error)
	GetList(companyID string, queryParam models.OrdersListRequest) (res models.OrderListResponse, err error)
	GetLocation(ID int) (models.Order, error)
	GetDetailedByPrimaryKey(ID int) (models.OrderShowResponse, error)
	Update(userID int64, entity *models.UpdateOrderRequest) (rowsAffected int64, err error)
	GetByStatus(companyID string, Status int) (order []models.Order, err error)
	GetByPhone(companyID string, Phone string) (models.Order, error)
	GetByPrimaryKey(ID int) (models.OrderShowResponse, error)
	Delete(entity models.DeleteOrderRequest) error
	SetOrderPrice(entity models.SetOrderPriceRequest) error
	// Payments: every change runs inside one transaction and ends with
	// order_payment_status_recalc; the returned summary is the new state.
	GetPaymentSummary(orderID int) (models.PaymentSummary, error)
	AddPayment(userID int64, entity models.AddOrderPaymentRequest) (models.PaymentSummary, error)
	DeleteTransaction(userID int64, transactionID int, companyID string) (models.PaymentSummary, error)
	GetOrderPayment(transactionID int) (models.OrderPaymentRef, error)
	RecalcPaymentStatus(orderID int) error
	AddComment(entity models.CreateOrderComment) error
	GetByUuid(uuid string) (models.OrderReceipt, error)
	// Band joylar: holati 6 (Topshirildi) bo'lmagan buyurtmalarning slots qiymatlari.
	GetBusySlots(companyID string) ([]models.BusySlot, error)
}
