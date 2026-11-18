package repo

import "bw-erp/models"

type UserI interface {
	Create(entity models.CreateUserModel) (int64, error)
	GetByPhone(phone string) (models.AuthUserModel, error)
	// GetById: faqat o'chirilmagan xodim (kirish, tahrirlash, o'chirish uchun).
	GetById(id int64) (models.User, error)
	// GetByIdWithDeleted: o'chirilgan xodim ham qaytadi (DeletedAt to'ldirilgan) — ko'rsatish uchun.
	GetByIdWithDeleted(id int64) (models.User, error)
	GetList(companyID string) ([]models.UserListItem, error)
	ChangePassword(userID int64, entity models.ChangePasswordRequest) error
	Edit(entity models.UpdateUserRequest) (rowsAffected int64, err error)
	GetCouriesList(companyID string) ([]models.GetCouriesResponse, error)
	// Delete: soft delete (deleted_at, deleted_by, is_active=false, chat_id=NULL) va user_companies.role_id = NULL.
	Delete(id int64, deletedBy int64) (int64, error)
}
