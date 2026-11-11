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

func (h *Handler) Create(c *gin.Context) {
	var body models.CreateUserModel

	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	// RequirePermission("user_create") tekshirgan korxona
	body.CompanyID = c.GetString("company_id")

	// joriy korxona so'rovdagi huquqlar ichida bo'lishi shart (korxonasiz xodim foydasiz)
	found := false
	for _, p := range body.Permissions {
		if p.CompanyID == body.CompanyID {
			found = true
		}
	}
	if !found {
		h.handleResponse(c, http.BadRequest, "company_id mismatch")
		return
	}

	// har bir korxonaga faqat o'sha korxonada user_create huquqi bilan xodim qo'shiladi
	for _, p := range body.Permissions {
		if !middleware.UserCan(c, p.CompanyID, "user_create") {
			h.handleResponse(c, http.PermissionDenied, "Permission denied")
			return
		}
	}

	id, err := h.Stg.User().Create(body)
	if errors.Is(err, models.ErrPhoneTaken) {
		h.handleResponse(c, http.Conflict, "Bu raqam band")
		return
	}
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, gin.H{"id": id})
}

func (h *Handler) GetList(c *gin.Context) {
	var body models.GetUserListRequest
	if err := c.ShouldBindQuery(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	users, err := h.Stg.User().GetList(body.CompanyID)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	h.handleResponse(c, http.OK, users)
}

func (h *Handler) Edit(c *gin.Context) {
	var body models.UpdateUserRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	userID, err := strconv.ParseInt(body.ID, 10, 64)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	// RequirePermission("edit_user") tekshirgan korxona so'rovdagi huquqlar ichida bo'lishi shart
	companyID := c.GetString("company_id")
	found := false
	for _, p := range body.Permissions {
		if p.CompanyID == companyID {
			found = true
		}
	}
	if !found {
		h.handleResponse(c, http.BadRequest, "company_id mismatch")
		return
	}

	// har bir korxona huquqini faqat o'sha korxonada edit_user huquqi bilan o'zgartiradi
	for _, p := range body.Permissions {
		if !middleware.UserCan(c, p.CompanyID, "edit_user") {
			h.handleResponse(c, http.PermissionDenied, "Permission denied")
			return
		}
	}

	// o'chirilgan xodim tahrirlanmaydi
	target, err := h.Stg.User().GetById(userID)
	if errors.Is(err, sql.ErrNoRows) {
		h.handleResponse(c, http.NOT_FOUND, "user not found")
		return
	}
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	// boshqa korxona xodimi tahrirlanmaydi
	inCompany := false
	for _, uc := range target.UserPermissionByCompany {
		if uc.CompanyID == companyID {
			inCompany = true
		}
	}
	if !inCompany {
		h.handleResponse(c, http.NOT_FOUND, "user not found")
		return
	}

	value, ok := c.Get("user")
	if !ok {
		h.handleResponse(c, http.Unauthorized, "user not found in context")
		return
	}
	current, ok := value.(models.User)
	if !ok {
		h.handleResponse(c, http.Unauthorized, "user not found in context")
		return
	}

	// super adminni faqat super admin tahrirlaydi
	if target.IsSuperAdmin && !current.IsSuperAdmin {
		h.handleResponse(c, http.PermissionDenied, "Permission denied")
		return
	}

	// chaqiruvchi edit_user ga ega korxonalar: permissions[] da yo'qlari xodimdan olib tashlanadi
	body.ManageableCompanyIDs = []string{}
	for _, uc := range current.UserPermissionByCompany {
		if middleware.UserCan(c, uc.CompanyID, "edit_user") {
			body.ManageableCompanyIDs = append(body.ManageableCompanyIDs, uc.CompanyID)
		}
	}

	_, err = h.Stg.User().Edit(body)
	if errors.Is(err, models.ErrPhoneTaken) {
		h.handleResponse(c, http.Conflict, "Bu raqam band")
		return
	}
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, true)
}

func (h *Handler) GetById(c *gin.Context) {
	userID := c.Param("user-id")

	// if !utils.IsValidUUID(userID) {
	// 	h.handleResponse(c, http.InvalidArgument, "user id is an invalid uuid")
	// 	return
	// }

	user_id, err := strconv.Atoi(userID)

	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	user, err := h.Stg.User().GetById(int64(user_id))

	if errors.Is(err, sql.ErrNoRows) {
		h.handleResponse(c, http.NOT_FOUND, "user not found")
		return
	}
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	h.handleResponse(c, http.OK, user)
}

func (h *Handler) GetCouriesList(c *gin.Context) {
	var req models.GetCouriesListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	couriers, err := h.Stg.User().GetCouriesList(req.CompanyID)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, couriers)
}

// DeleteUser: soft delete. Korxona RequirePermission("delete_user") tekshirgan query company_id.
func (h *Handler) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user-id"), 10, 64)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	value, ok := c.Get("user")
	if !ok {
		h.handleResponse(c, http.Unauthorized, "user not found in context")
		return
	}
	current, ok := value.(models.User)
	if !ok {
		h.handleResponse(c, http.Unauthorized, "user not found in context")
		return
	}
	if current.ID == userID {
		h.handleResponse(c, http.BadRequest, "O'zini o'chirib bo'lmaydi")
		return
	}

	target, err := h.Stg.User().GetById(userID)
	if errors.Is(err, sql.ErrNoRows) {
		h.handleResponse(c, http.NOT_FOUND, "user not found")
		return
	}
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	companyID := c.GetString("company_id")
	inCompany := false
	for _, uc := range target.UserPermissionByCompany {
		if uc.CompanyID == companyID {
			inCompany = true
		}
	}
	if !inCompany {
		h.handleResponse(c, http.NOT_FOUND, "user not found")
		return
	}

	// super adminni faqat super admin o'chiradi
	if target.IsSuperAdmin && !current.IsSuperAdmin {
		h.handleResponse(c, http.PermissionDenied, "Permission denied")
		return
	}

	_, err = h.Stg.User().Delete(userID, current.ID)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	h.handleResponse(c, http.OK, true)
}
