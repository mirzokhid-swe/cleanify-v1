package models

import "encoding/json"

type CreateCompanyModel struct {
	Name    string `json:"name" binding:"required" minLength:"2" maxLength:"255"`
	OwnerId int64  `json:"owner_id" binding:"required"`
}

type Company struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CompanyListItem is one row of the super admin companies list.
type CompanyListItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsersCount int    `json:"users_count"`
	IsBlocked  bool   `json:"is_blocked"`
}

// CompanyShow is the company page: list fields plus every setting (defaults filled in).
type CompanyShow struct {
	CompanyListItem
	Settings map[string]json.RawMessage `json:"settings"`
}
