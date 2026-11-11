package api

import (
	"bw-erp/api/handlers"
	"bw-erp/api/middleware"
	"bw-erp/config"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetUpRouter(h handlers.Handler, cfg config.Config) (r *gin.Engine) {
	r = gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	// Har qanday xatolik (>=400, panic) Telegram guruhga (ERROR_BOT_TOKEN, ERROR_CHAT_ID).
	r.Use(middleware.ErrorReport(cfg.Release_Mode, middleware.TelegramSender(cfg.ErrorBotToken, cfg.ErrorChatID)))
	if cfg.ErrorBotToken != "" && cfg.ErrorChatID != "" {
		log.Println("[error-report] telegram yoqiq, chat", cfg.ErrorChatID)
	} else {
		log.Println("[error-report] telegram o'chiq: ERROR_BOT_TOKEN yoki ERROR_CHAT_ID bo'sh")
	}

	r.Use(customCORSMiddleware())

	r.GET("api/ping", h.Ping)
	r.GET("api/get-order-receipt-by-uuid/:uuid", h.GetOrderReceiptByUuid)

	baseRouter := r.Group("/api/v1")
	{
		{
			authRouter := baseRouter.Group("/auth")
			// authRouter.POST("/login", h.AuthUser)
			// Middleware har yo'lga alohida: group.Use to'planib, refresh-token ga ham
			// token tekshiruvini qo'shib qo'yardi (muddati o'tgan token bilan refresh 403).
			authRouter.POST("/me", middleware.UserActiveMiddleware(h.Stg), h.CurrentUser)
			authRouter.POST("/refresh-token", h.RefreshToken)
			authRouter.POST("/change-password", middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg), h.ChangePassword)
		}

		notificationRouter := baseRouter.Group("/notifications")
		{
			notificationRouter.GET("/ws", h.HandleNotificationWebSocket) // WebSocket endpoint

			// Web Push endpoints
			webpushRouter := notificationRouter.Group("/webpush")
			webpushRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg))
			{
				webpushRouter.POST("/subscribe", h.SavePushSubscription)
				// webpushRouter.GET("/subscription", h.GetPushSubscription)
				// webpushRouter.DELETE("/subscription", h.DeletePushSubscription)
				// webpushRouter.GET("/subscriptions", h.GetAllPushSubscriptions)
				// webpushRouter.POST("/send", h.SendPushNotification)
			}

			notificationRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("list", h.GetMyNotifications)
			notificationRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("/unread-notifications-count", h.UnreadNotificationsCount)
		}

		usersRoutes(baseRouter.Group("/users"), &h, middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg))
		rolesRoutes(baseRouter.Group("/roles"), &h, middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg))
	}

	baseRouter.Static("/uploads", "./uploads")

	{
		companyRoutes(baseRouter.Group("/company"), &h, middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg))
	}

	{
		orderRouter := baseRouter.Group("orders")
		orderRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).POST("", h.CreateOrderModel)
		orderRouter.GET("", middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg), middleware.RequirePermission("get_orders_list"), h.GetOrdersList)
		orderRouter.GET("/busy-slots", middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg), h.GetBusySlots)
		orderRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("/:order-id", h.GetOrderByPrimaryKey)
		orderRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).POST("/edit", h.UpdateOrderModel)
		orderRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).POST("/set-price", h.SetOrderPrice)
		orderRouter.POST("add-payment", middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg), middleware.RequirePermission("add_payment"), h.AddOrderPayment)
		// huquq handler'da: delete_transaction yoki o'zi kiritgan to'lov
		orderRouter.DELETE("/transactions/:id", middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg), h.DeleteOrderTransaction)
		orderRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("get-transactions-by-order", h.GetTransactionByOrder)
		orderRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).DELETE("", h.DeleteOrder)
		orderRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).POST("/comment", h.AddOrderComment)
	}

	{
		notificationSettingRouter := baseRouter.Group("notification-setting")
		notificationSettingRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).POST("", h.SetNotificationSetting)
		notificationSettingRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("", h.UsersListForNotificationSettings)
		notificationSettingRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("get-users-by-status", h.GetUsersByStatus)
	}

	{
		orderStatuses := baseRouter.Group("order-statuses")
		orderStatuses.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("", h.GetOrderStatusesList) //
		orderStatuses.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).PUT("", h.UpdateOrderStatusModel)
		orderStatuses.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("/get-by-primary-key/:id", h.GetOrderStatusById)
		orderStatuses.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).PUT("/reorder", h.ReorderOrderStatus)
	}

	{
		orderItemRoutes(baseRouter.Group("order-items"), &h, middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg))
	}

	{
		orderItemTypeRoutes(baseRouter.Group("order-item-type"), &h, middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg))
	}

	{
		statistics := baseRouter.Group("statistics")
		statistics.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("work-volume", h.GetWorkVolumeList) //
		statistics.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("get-service-statistics-payment", h.GetServicePaymentStatistics)
	}
	{
		statistics := baseRouter.Group("permissions")
		statistics.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("", h.GetPermissionList)
	}

	{
		clientRouter := baseRouter.Group("/client")
		clientRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).POST("", h.CreateClientModel) //
		clientRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("/get-by-primary-key/:client-id", h.GetClientByPrimaryKey)
		clientRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("", h.GetClientsList) //
		clientRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("/set-location/:client-id", h.SetLocation)
		clientRouter.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).PUT("", h.UpdateClient) //
	}

	{
		telegramGroup := baseRouter.Group("/telegram-group")
		telegramGroup.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).POST("/verification", h.VerificationGroup) //
		telegramGroup.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("", h.GetTelegramGroupList)            //
		telegramGroup.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).GET("/get-by-primary-key/:id", h.GetTelegramGroupByPrimaryKey)
		telegramGroup.Use(middleware.AuthMiddleware(), middleware.UserActiveMiddleware(h.Stg)).PUT("/:id", h.UpdateTelegramGroup)
	}

	return
}

func customCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		// Add WebSocket specific headers
		if c.Request.Header.Get("Upgrade") == "websocket" {
			c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, Connection, Upgrade, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Extensions")
		} else {
			c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// orderItemTypeRoutes registers the order item type routes. auth runs before
// RequirePermission on every route and must leave a models.User in the context
// under "user" (AuthMiddleware + UserActiveMiddleware in production).
func orderItemTypeRoutes(g *gin.RouterGroup, h *handlers.Handler, auth ...gin.HandlerFunc) {
	with := func(slug string, handler gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, auth...), middleware.RequirePermission(slug), handler)
	}
	g.POST("", with("create_order_item_type", h.CreateOrderItemTypeModel)...)
	g.GET("", with("get_order_item_type", h.GetOrderItemTypesByCompany)...)
	g.PUT("", with("edit_order_item_type", h.UpdateOrderItemType)...)
	g.GET("get-by-primary-key/:id", with("get_order_item_type", h.GetOrderItemTypeByID)...)
	g.DELETE("/:id", with("delete_order_item_type", h.DeleteOrderItemType)...)
}

// orderItemRoutes registers the order item routes. auth must leave a
// models.User in the context under "user". The company comes from the order,
// so each handler checks the permission itself with middleware.UserCan.
func orderItemRoutes(g *gin.RouterGroup, h *handlers.Handler, auth ...gin.HandlerFunc) {
	with := func(handler gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, auth...), handler)
	}
	g.POST("", with(h.CreateOrderItemModel)...)
	g.POST("edit", with(h.UpdateOrderItemModel)...)
	g.DELETE("/:id", with(h.DeleteOrderItemByID)...)
	g.POST("/edit-status", with(h.UpdateOrderItemStatus)...)
}

// companyRoutes registers the super admin company routes. auth must leave a
// models.User in the context under "user"; SuperAdminMiddleware then checks it.
func companyRoutes(g *gin.RouterGroup, h *handlers.Handler, auth ...gin.HandlerFunc) {
	with := func(handler gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, auth...), middleware.SuperAdminMiddleware(), handler)
	}
	g.POST("", with(h.CreateCompanyModel)...)
	g.GET("", with(h.GetCompaniesList)...)
	g.GET("/:id", with(h.GetCompanyShow)...)
	g.PATCH("/:id/settings", with(h.UpdateCompanySetting)...)
	g.POST("/:id/block", with(h.BlockCompany)...)
	g.POST("/:id/unblock", with(h.UnblockCompany)...)
}

// usersRoutes registers the user routes. auth must leave a models.User in the
// context under "user". Write routes need a permission slug in the request's
// company_id; read routes stay auth only.
func usersRoutes(g *gin.RouterGroup, h *handlers.Handler, auth ...gin.HandlerFunc) {
	with := func(slug string, handler gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, auth...), middleware.RequirePermission(slug), handler)
	}
	authOnly := func(handler gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, auth...), handler)
	}
	g.POST("", with("user_create", h.Create)...)
	g.POST("/edit", with("edit_user", h.Edit)...)
	g.DELETE("/:user-id", with("delete_user", h.DeleteUser)...)
	g.GET("", authOnly(h.GetList)...)
	g.GET("/:user-id", authOnly(h.GetById)...)
	g.GET("/couriers", authOnly(h.GetCouriesList)...)

	g.POST("/employees", authOnly(h.CreateEmployee)...)
	g.GET("/employees", authOnly(h.GetEmployeeList)...)
	g.GET("/employees/show", authOnly(h.ShowEmployeeDetailedData)...)
	g.POST("/employees/add-transaction", authOnly(h.AddTransaction)...)
	g.POST("/employees/attendance", authOnly(h.Attendance)...)
}

// rolesRoutes registers the role routes. auth must leave a models.User in the
// context under "user". Reads check one of several slugs in the handler.
func rolesRoutes(g *gin.RouterGroup, h *handlers.Handler, auth ...gin.HandlerFunc) {
	with := func(slug string, handler gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, auth...), middleware.RequirePermission(slug), handler)
	}
	authOnly := func(handler gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, auth...), handler)
	}
	g.GET("", authOnly(h.GetRoles)...)
	g.GET("/:id", authOnly(h.GetRole)...)
	g.POST("", with("manage_roles", h.CreateRole)...)
	g.PUT("/:id", with("manage_roles", h.UpdateRole)...)
	g.DELETE("/:id", with("manage_roles", h.DeleteRole)...)
}
