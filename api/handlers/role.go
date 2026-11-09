package handlers

import (
	"bw-erp/api/http"
	"bw-erp/api/middleware"
	"bw-erp/models"
	"database/sql"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// canReadRoles: rollar ro'yxati xodim formasi va rollar sahifasi uchun kerak.
func canReadRoles(c *gin.Context, companyID string) bool {
	for _, slug := range []string{"get_users_list", "user_create", "edit_user", "manage_roles"} {
		if middleware.UserCan(c, companyID, slug) {
			return true
		}
	}
	return false
}

func (h *Handler) GetRoles(c *gin.Context) {
	companyID := c.Query("company_id")
	if !canReadRoles(c, companyID) {
		h.handleResponse(c, http.PermissionDenied, nil)
		return
	}

	roles, err := h.Stg.Role().GetList(companyID)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	h.handleResponse(c, http.OK, roles)
}

func (h *Handler) GetRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	companyID := c.Query("company_id")
	if !canReadRoles(c, companyID) {
		h.handleResponse(c, http.PermissionDenied, nil)
		return
	}

	role, err := h.Stg.Role().GetByID(id, companyID)
	if errors.Is(err, sql.ErrNoRows) {
		h.handleResponse(c, http.NOT_FOUND, "role not found")
		return
	}
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	h.handleResponse(c, http.OK, role)
}

func (h *Handler) CreateRole(c *gin.Context) {
	var body models.RoleInput
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	body.CompanyID = c.GetString("company_id")

	id, err := h.Stg.Role().Create(body)
	if errors.Is(err, models.ErrRoleNameTaken) {
		h.handleResponse(c, http.Conflict, "Bu nom band")
		return
	}
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	h.handleResponse(c, http.Created, gin.H{"id": id})
}

func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	var body models.RoleInput
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	body.CompanyID = c.GetString("company_id")

	rows, err := h.Stg.Role().Update(id, body)
	if errors.Is(err, models.ErrRoleNameTaken) {
		h.handleResponse(c, http.Conflict, "Bu nom band")
		return
	}
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	if rows == 0 {
		h.handleResponse(c, http.NOT_FOUND, "role not found")
		return
	}
	h.handleResponse(c, http.OK, true)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	companyID := c.GetString("company_id")

	rows, err := h.Stg.Role().Delete(id, companyID)
	if errors.Is(err, models.ErrRoleInUse) {
		role, err := h.Stg.Role().GetByID(id, companyID)
		if err != nil {
			h.handleResponse(c, http.InternalServerError, err.Error())
			return
		}
		h.handleResponse(c, http.Conflict, gin.H{"users_count": role.UsersCount})
		return
	}
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	if rows == 0 {
		h.handleResponse(c, http.NOT_FOUND, "role not found")
		return
	}
	h.handleResponse(c, http.OK, true)
}
