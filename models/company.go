package models

import (
	"encoding/json"
	"errors"
)

type CreateCompanyModel struct {
	Name string `json:"name" binding:"required"`
}

var ErrCompanyNameTaken = errors.New("company name taken")

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
