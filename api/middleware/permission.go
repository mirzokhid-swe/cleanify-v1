package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"bw-erp/api/http"
	"bw-erp/models"

	"github.com/gin-gonic/gin"
)

// CompanyIDFromRequest returns the company_id of the request: the query
// parameter when present, otherwise the "company_id" field of the JSON body.
func CompanyIDFromRequest(c *gin.Context) string {
	if id := c.Query("company_id"); id != "" {
		return id
	}
	return bodyCompanyID(c)
}

// bodyCompanyID reads "company_id" from the JSON body and restores the body
// so the handler can bind it again.
func bodyCompanyID(c *gin.Context) string {
	if c.Request == nil || c.Request.Body == nil {
		return ""
	}
	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(b))
	var body struct {
		CompanyID string `json:"company_id"`
	}
	if err := json.Unmarshal(b, &body); err != nil {
		return ""
	}
	return body.CompanyID
}

func companyPermissions(user models.User, companyID string) (models.UserPermissionByCompany, bool) {
	for _, uc := range user.UserPermissionByCompany {
		if uc.CompanyID == companyID {
			return uc, true
		}
	}
	return models.UserPermissionByCompany{}, false
}

func hasPermission(can string, slug string) bool {
	for _, part := range strings.Split(can, "|") {
		if part == slug {
			return true
		}
	}
	return false
}

func forbid(c *gin.Context, description string) {
	c.AbortWithStatusJSON(http.PermissionDenied.Code, http.Response{
		Status:      http.PermissionDenied.Status,
		Description: description,
		Data:        nil,
	})
}

func locked(c *gin.Context) {
	c.AbortWithStatusJSON(http.Locked.Code, http.Response{
		Status:      http.Locked.Status,
		Description: http.Locked.Description,
		Data:        nil,
	})
}

// UserCan reports whether the user in the context has slug in companyID.
// For handlers that learn the company from the resource, not the request.
func UserCan(c *gin.Context, companyID string, slug string) bool {
	value, ok := c.Get("user")
	if !ok {
		return false
	}
	user, ok := value.(models.User)
	if !ok {
		return false
	}
	uc, found := companyPermissions(user, companyID)
	return found && !uc.IsBlocked && hasPermission(uc.Can, slug)
}

// RequirePermission allows the request only when the user set by
// UserActiveMiddleware has the permission slug in the company named by the
// request (query or JSON body company_id; when both are sent they must be
// equal). On success the verified company id is stored in the context under
// "company_id" and handlers must use it instead of re-reading the request.
func RequirePermission(slug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, ok := c.Get("user")
		if !ok {
			forbid(c, "Permission denied")
			return
		}
		user, ok := value.(models.User)
		if !ok {
			forbid(c, "Permission denied")
			return
		}
		queryID, bodyID := c.Query("company_id"), bodyCompanyID(c)
		if queryID != "" && bodyID != "" && queryID != bodyID {
			forbid(c, "company_id mismatch")
			return
		}
		companyID := queryID
		if companyID == "" {
			companyID = bodyID
		}
		if companyID == "" {
			forbid(c, "company_id is required")
			return
		}
		uc, found := companyPermissions(user, companyID)
		if !found {
			forbid(c, "Permission denied")
			return
		}
		if uc.IsBlocked {
			locked(c)
			return
		}
		if !hasPermission(uc.Can, slug) {
			forbid(c, "Permission denied")
			return
		}
		c.Set("company_id", companyID)
		c.Next()
	}
}
