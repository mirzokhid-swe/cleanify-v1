package handlers

import (
	"database/sql"
	"errors"

	"bw-erp/api/http"
	"bw-erp/models"
	"bw-erp/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) CreateCompanyModel(c *gin.Context) {
	var body models.CreateCompanyModel
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	_, err := h.Stg.User().GetById(body.OwnerId)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	id := uuid.New()

	err = h.Stg.Company().Create(id.String(), body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.Created, id)
}

func (h *Handler) GetCompanyByOwnerId(c *gin.Context) {
	err := utils.TokenValid(c)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	jwtData, err := utils.ExtractTokenID(c)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	user, err := h.Stg.User().GetById(jwtData.UserID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	company, err := h.Stg.Company().GetByOwnerId(user.ID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, company)
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
