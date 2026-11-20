package repo

import (
	"encoding/json"

	"bw-erp/models"
)

type CompanyStorageI interface {
	// Create writes the company with its default statuses, item types, roles and
	// payment purposes and attaches userID as Admin, all in one transaction.
	// Returns models.ErrCompanyNameTaken when the name is in use.
	Create(id string, name string, userID int64) error
	GetById(id string) (models.Company, error)
	GetList() ([]models.CompanyListItem, error)
	// GetShow returns sql.ErrNoRows when the company does not exist.
	GetShow(id string) (models.CompanyShow, error)
	UpsertSetting(companyID string, key string, value json.RawMessage, userID int64) error
	SetBlocked(companyID string, blocked bool, userID int64) error
}
