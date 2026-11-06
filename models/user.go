package models

import (
	"database/sql"
	"encoding/json"
)

type CreateUserModel struct {
	Fullname    string `json:"fullname" binding:"required" minLength:"2" maxLength:"255" example:"John Doe"`
	Phone       string `json:"phone" binding:"required" example:"991234567"`
	CompanyID   string `json:"company_id" binding:"required"`
	Permissions []struct {
		CompanyID     string `json:"company_id" binding:"required"`
		PermissionIDs []int  `json:"permission_ids"`
		IsCourier     bool   `json:"is_courier" default:"false"`
	} `json:"permissions"`
}

type AuthUserModel struct {
	ID       string `json:"id"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID                      int64                     `json:"id"`
	Fullname                string                    `json:"fullname"`
	Phone                   string                    `json:"phone"`
	CompanyID               *string                   `json:"company_id"`
	IsActive                bool                      `json:"is_active"`
	IsSuperAdmin            bool                      `json:"is_super_admin"`
	UserPermissionByCompany []UserPermissionByCompany `json:"user_permissions_by_company"`
}

type GetCouriesResponse struct {
	ID       string `json:"id"`
	Fullname string `json:"fullname"`
}

type GetCouriesListRequest struct {
	CompanyID string `json:"company_id" form:"company_id" binding:"required"`
}

type UserPermissionByCompany struct {
	CompanyID     string          `json:"company_id"`
	CompanyName   string          `json:"company_name"`
	PermissionIDs []int           `json:"permission_ids"`
	Can           string          `json:"can"`
	IsCourier     bool            `json:"is_courier"`
	IsBlocked     bool            `json:"is_blocked"`
	Settings      json.RawMessage `json:"settings"`
}

type UpdateUserRequest struct {
	ID          string `json:"id" binding:"required"`
	Fullname    string `json:"fullname"`
	Phone       string `json:"phone"`
	IsActive    *bool  `json:"is_active"`
	Permissions []struct {
		CompanyID     string `json:"company_id" binding:"required"`
		PermissionIDs []int  `json:"permission_ids"`
		IsCourier     bool   `json:"is_courier" default:"false"`
	} `json:"permissions"`
}

type GetUserListRequest struct {
	CompanyID string `json:"company_id" form:"company_id" binding:"required"`
}

type NullString struct {
	sql.NullString
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}
