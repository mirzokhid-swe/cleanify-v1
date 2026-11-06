package middleware

import (
	"bw-erp/models"

	"github.com/gin-gonic/gin"
)

// SuperAdminMiddleware allows only users with users.is_super_admin = true.
// It needs the user set by UserActiveMiddleware; the check does not depend on a company.
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, ok := c.Get("user")
		if !ok {
			forbid(c, "Permission denied")
			return
		}
		user, ok := value.(models.User)
		if !ok || !user.IsSuperAdmin {
			forbid(c, "Permission denied")
			return
		}
		c.Next()
	}
}
