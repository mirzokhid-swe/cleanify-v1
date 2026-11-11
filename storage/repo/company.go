package repo

import (
	"encoding/json"

	"bw-erp/models"
)

type CompanyStorageI interface {
	Create(id string, entity models.CreateCompanyModel) error
	GetByOwnerId(ownerId int64) ([]models.Company, error)
	GetById(id string) (models.Company, error)
	GetList() ([]models.CompanyListItem, error)
	// GetShow returns sql.ErrNoRows when the company does not exist.
	GetShow(id string) (models.CompanyShow, error)
	UpsertSetting(companyID string, key string, value json.RawMessage, userID int64) error
	SetBlocked(companyID string, blocked bool, userID int64) error
}
