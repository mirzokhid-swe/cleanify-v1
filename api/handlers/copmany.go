package handlers

import (
	"database/sql"
	"errors"
	"strings"

	"bw-erp/api/http"
	"bw-erp/models"
	"bw-erp/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateCompanyModel: super admin yangi korxona qo'shadi; standart holatlar,
// turlar, rollar va to'lov maqsadlari bilan birga yaratiladi.
func (h *Handler) CreateCompanyModel(c *gin.Context) {
	var body models.CreateCompanyModel
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		h.handleResponse(c, http.BadRequest, "Korxona nomini kiriting")
		return
	}

	id := uuid.New().String()
	err := h.Stg.Company().Create(id, name, currentUserID(c))
	if errors.Is(err, models.ErrCompanyNameTaken) {
		h.handleResponse(c, http.Conflict, "Bu nomli korxona bor")
		return
	}
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, gin.H{"id": id})
}

// currentUserID returns the id of the user set by UserActiveMiddleware (0 when absent).
func currentUserID(c *gin.Context) int64 {
	if value, ok := c.Get("user"); ok {
		if user, ok := value.(models.User); ok {
			return user.ID
		}
	}
	return 0
}

func (h *Handler) GetCompaniesList(c *gin.Context) {
	data, err := h.Stg.Company().GetList()
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	h.handleResponse(c, http.OK, data)
}

func (h *Handler) GetCompanyShow(c *gin.Context) {
	id := c.Param("id")
	if !utils.IsValidUUID(id) {
		h.handleResponse(c, http.InvalidArgument, "id is an invalid uuid")
		return
	}
	data, err := h.Stg.Company().GetShow(id)
	if errors.Is(err, sql.ErrNoRows) {
		h.handleResponse(c, http.NOT_FOUND, "Company not found!")
		return
	}
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	h.handleResponse(c, http.OK, data)
}

// UpdateCompanySetting saves one setting and returns the full settings map.
func (h *Handler) UpdateCompanySetting(c *gin.Context) {
	id := c.Param("id")
	if !utils.IsValidUUID(id) {
		h.handleResponse(c, http.InvalidArgument, "id is an invalid uuid")
		return
	}
	var body models.UpdateCompanySettingRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	value, err := models.ValidateSetting(body.Key, body.Value)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if _, err := h.Stg.Company().GetShow(id); errors.Is(err, sql.ErrNoRows) {
		h.handleResponse(c, http.NOT_FOUND, "Company not found!")
		return
	} else if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if err := h.Stg.Company().UpsertSetting(id, body.Key, value, currentUserID(c)); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	show, err := h.Stg.Company().GetShow(id)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	h.handleResponse(c, http.OK, show.Settings)
}

func (h *Handler) BlockCompany(c *gin.Context)   { h.setCompanyBlocked(c, true) }
func (h *Handler) UnblockCompany(c *gin.Context) { h.setCompanyBlocked(c, false) }

func (h *Handler) setCompanyBlocked(c *gin.Context, blocked bool) {
	id := c.Param("id")
	if !utils.IsValidUUID(id) {
		h.handleResponse(c, http.InvalidArgument, "id is an invalid uuid")
		return
	}
	if _, err := h.Stg.Company().GetShow(id); errors.Is(err, sql.ErrNoRows) {
		h.handleResponse(c, http.NOT_FOUND, "Company not found!")
		return
	} else if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if err := h.Stg.Company().SetBlocked(id, blocked, currentUserID(c)); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	show, err := h.Stg.Company().GetShow(id)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	h.handleResponse(c, http.OK, show)
}
