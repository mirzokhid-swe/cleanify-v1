package models

import "errors"

var ErrRoleNameTaken = errors.New("role name taken")
var ErrRoleInUse = errors.New("role in use")
var ErrRoleForeign = errors.New("role belongs to another company")

type Role struct {
	ID            int64  `json:"id"`
	CompanyID     string `json:"company_id"`
	Name          string `json:"name"`
	PermissionIDs []int  `json:"permission_ids"`
	UsersCount    int    `json:"users_count"`
}

type RoleInput struct {
	CompanyID     string `json:"company_id" binding:"required"`
	Name          string `json:"name" binding:"required,min=2,max=100"`
	PermissionIDs []int  `json:"permission_ids"`
}
