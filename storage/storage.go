package storage

import (
	"bw-erp/storage/postgres"
	"bw-erp/storage/repo"

	"github.com/jmoiron/sqlx"
)

type StorageI interface {
	Client() repo.ClientStorageI
	BotUser() repo.BotUserI
	Company() repo.CompanyStorageI
	OrderItemType() repo.OrderItemTypeI
	OrderItem() repo.OrderItemI
	Order() repo.OrderI
	Permission() repo.PermissionI
	Role() repo.RoleI
	Statistics() repo.StatisticsI
	TelegramBot() repo.TelegramBotI
	TelegramGroup() repo.TelegramGroupI
	TelegramSession() repo.TelegramSessionI
	User() repo.UserI
	OrderStatus() repo.OrderStatusI
	StatusChangeHistory() repo.StatusChangeHistoryI
	Employee() repo.EmployeeI
	NotificationSetting() repo.NotificationSettingI
	Notification() repo.NotificationI
	Comment() repo.CommentStorageI
	WebPush() repo.WebPushStorageI
}

type storagePg struct {
	db                  *sqlx.DB
	client              repo.ClientStorageI
	botUser             repo.BotUserI
	user                repo.UserI
	company             repo.CompanyStorageI
	orderItemType       repo.OrderItemTypeI
	order               repo.OrderI
	orderItem           repo.OrderItemI
	permission          repo.PermissionI
	role                repo.RoleI
	statistics          repo.StatisticsI
	telegramGroup       repo.TelegramGroupI
	telegramSession     repo.TelegramSessionI
	telegramBot         repo.TelegramBotI
	orderStatus         repo.OrderStatusI
	statusChangeHistory repo.StatusChangeHistoryI
	employee            repo.EmployeeI
	notificationSetting repo.NotificationSettingI
	notification        repo.NotificationI
	comment             repo.CommentStorageI
	webPush             repo.WebPushStorageI
}

func NewStoragePg(db *sqlx.DB) StorageI {
	return &storagePg{
		db:                  db,
		client:              postgres.NewClientRepo(db),
		botUser:             postgres.NewBotUserRepo(db),
		user:                postgres.NewUserRepo(db),
		company:             postgres.NewCompanyRepo(db),
		orderItemType:       postgres.NewOrderItemTypeRepo(db),
		order:               postgres.NewOrderRepo(db),
		orderItem:           postgres.NewOrderItemRepo(db),
		permission:          postgres.NewPermissionRepo(db),
		role:                postgres.NewRoleRepo(db),
		statistics:          postgres.NewStatisticsRepo(db),
		telegramGroup:       postgres.NewTelegramGroupRepo(db),
		telegramSession:     postgres.NewTelegramSessionRepo(db),
		orderStatus:         postgres.NewOrderStatusRepo(db),
		statusChangeHistory: postgres.NewStatusChangeHistoryRepo(db),
		employee:            postgres.NewEmployeeRepo(db),
		notificationSetting: postgres.NewNotificationSettingRepo(db),
		notification:        postgres.NewNotificationRepo(db),
		comment:             postgres.NewCommentRepo(db),
		webPush:             postgres.NewWebPushRepo(db),
	}
}

func (s storagePg) Client() repo.ClientStorageI {
	return s.client
}

func (s storagePg) BotUser() repo.BotUserI {
	return s.botUser
}

func (s storagePg) Company() repo.CompanyStorageI {
	return s.company
}

func (s storagePg) OrderItemType() repo.OrderItemTypeI {
	return s.orderItemType
}

func (s storagePg) OrderItem() repo.OrderItemI {
	return s.orderItem
}

func (s storagePg) Order() repo.OrderI {
	return s.order
}

func (s storagePg) Permission() repo.PermissionI {
	return s.permission
}

func (s storagePg) Role() repo.RoleI {
	return s.role
}

func (s storagePg) Statistics() repo.StatisticsI {
	return s.statistics
}

func (s storagePg) TelegramBot() repo.TelegramBotI {
	return s.telegramBot
}

func (s storagePg) TelegramGroup() repo.TelegramGroupI {
	return s.telegramGroup
}

func (s storagePg) TelegramSession() repo.TelegramSessionI {
	return s.telegramSession
}

func (s storagePg) User() repo.UserI {
	return s.user
}

func (s storagePg) OrderStatus() repo.OrderStatusI {
	return s.orderStatus
}

func (s storagePg) StatusChangeHistory() repo.StatusChangeHistoryI {
	return s.statusChangeHistory
}

func (s storagePg) Employee() repo.EmployeeI {
	return s.employee
}

func (s storagePg) NotificationSetting() repo.NotificationSettingI {
	return s.notificationSetting
}

func (s storagePg) Notification() repo.NotificationI {
	return s.notification
}

func (s storagePg) Comment() repo.CommentStorageI {
	return s.comment
}

func (s *storagePg) WebPush() repo.WebPushStorageI {
	return s.webPush
}
