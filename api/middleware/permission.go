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
// The body is read and restored so the handler can bind it again.
func CompanyIDFromRequest(c *gin.Context) string {
	if id := c.Query("company_id"); id != "" {
		return id
	}
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
	c.AbortWithStatusJSON(http.Forbidden.Code, http.Response{
		Status:      http.Forbidden.Status,
		Description: description,
		Data:        nil,
	})
}

// RequirePermission allows the request only when the user set by
// UserActiveMiddleware has the permission slug in the company named by the
// request (query or JSON body company_id). On success the company id is
// stored in the context under "company_id".
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
		companyID := CompanyIDFromRequest(c)
		if companyID == "" {
			forbid(c, "company_id is required")
			return
		}
		uc, found := companyPermissions(user, companyID)
		if !found || !hasPermission(uc.Can, slug) {
			forbid(c, "Permission denied")
			return
		}
		c.Set("company_id", companyID)
		c.Next()
	}
}
