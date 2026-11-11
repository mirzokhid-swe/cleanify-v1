package repo

import "bw-erp/models"

type RoleI interface {
	GetList(companyID string) ([]models.Role, error)
	GetByID(id int64, companyID string) (models.Role, error)
	Create(in models.RoleInput) (int64, error)
	Update(id int64, in models.RoleInput) (int64, error)
	Delete(id int64, companyID string) (int64, error)
}
