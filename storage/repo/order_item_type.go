package repo

import "bw-erp/models"

type OrderItemTypeI interface {
	Create(id string, entity models.OrderItemTypeModel) error
	GetByCompany(CompanyID string) ([]models.OrderItemByCompany, error)
	Update(entity models.EditOrderItemTypeRequest) (rowsAffected int64, err error)
	// GetById returns the type; when companyID is not empty the type must belong to it.
	GetById(orderItemTypeId string, companyID string) (models.OrderItemByCompany, error)
	Delete(id string, companyID string) (rowsAffected int64, err error)
}
