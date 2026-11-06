package postgres

import (
	"bw-erp/helper"
	"bw-erp/models"
	"bw-erp/pkg/utils"
	"bw-erp/storage/repo"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) repo.UserI {
	return &userRepo{db: db}
}

func (stg userRepo) Create(entity models.CreateUserModel) error {
	var id string
	err := stg.db.QueryRow(`INSERT INTO users(
		phone,
		fullname,
		company_id,
		is_active
	) VALUES (
		$1,
		$2,
		$3,
		true
	) RETURNING id`,
		entity.Phone,
		entity.Fullname,
		entity.CompanyID,
	).Scan(&id)

	if err != nil {
		return err
	}

	// Add permissions if provided
	if len(entity.Permissions) > 0 {
		for _, permission := range entity.Permissions {
			// Check if user already has permissions for this company
			var userCompanyID string
			stg.db.QueryRow(`SELECT company_id FROM user_companies WHERE company_id = $1 AND user_id = $2`,
				permission.CompanyID, id).Scan(&userCompanyID)

			PermissionIDs := utils.SetArray(utils.IntSliceToInterface(permission.PermissionIDs))

			if len(userCompanyID) > 0 {
				// Update existing permissions
				query := `UPDATE "user_companies" SET 
					permission_ids = :permission_ids, 
					is_courier = :is_courier, 
					updated_at = now() 
				WHERE company_id = :company_id AND user_id = :user_id`

				permissionEditParams := map[string]interface{}{
					"permission_ids": PermissionIDs,
					"is_courier":     permission.IsCourier,
					"company_id":     permission.CompanyID,
					"user_id":        id,
				}

				query, arr := helper.ReplaceQueryParams(query, permissionEditParams)

				_, err := stg.db.Exec(query, arr...)

				if err != nil {
					return err
				}
			} else {
				// Insert new permissions
				_, err := stg.db.Exec(`INSERT INTO user_companies(
					permission_ids,
					company_id,
					user_id,
					is_courier
				) VALUES (
					$1,
					$2,
					$3,
					$4
				)`,
					PermissionIDs,
					permission.CompanyID,
					id,
					permission.IsCourier,
				)

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (stg userRepo) GetByPhone(phone string) (models.AuthUserModel, error) {
	var user models.AuthUserModel
	err := stg.db.QueryRow(`SELECT id, phone, password from users where phone = $1`, phone).Scan(
		&user.ID,
		&user.Phone,
		&user.Password,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (stg userRepo) GetById(id int64) (models.User, error) {
	var user models.User
	err := stg.db.QueryRow(`select u.id, u.fullname, u.phone, c.id, u.is_active, u.is_super_admin from users u left join companies c on c.id::text = u.company_id::text where u.id = $1`, id).Scan(
		&user.ID,
		&user.Fullname,
		&user.Phone,
		&user.CompanyID,
		&user.IsActive,
		&user.IsSuperAdmin,
	)
	if err != nil {
		return user, err
	}

	allPermissions, err := stg.GetAllPermissions()
	if err != nil {
		return user, err
	}

	rows, err := stg.db.Query(`select c.id, c.name, uc.permission_ids, uc.is_courier, c.is_blocked,
								coalesce((select jsonb_object_agg(cs.key, cs.value) from company_settings cs where cs.company_id = c.id), '{}'::jsonb)
								from user_companies uc
								inner join companies c on c.id::text = uc.company_id::text
								where uc.user_id = $1 order by c.name desc`, user.ID)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	for rows.Next() {
		var permissions models.UserPermissionByCompany
		var permissionIDsStr string
		var settings []byte
		if err := rows.Scan(&permissions.CompanyID, &permissions.CompanyName, &permissionIDsStr, &permissions.IsCourier, &permissions.IsBlocked, &settings); err != nil {
			return user, err
		}
		permissions.Settings = json.RawMessage(settings)

		if permissionIDsStr != "" {
			Permissions := utils.GetArray(permissionIDsStr)
			Permission := ""
			permissions.PermissionIDs = utils.InterfaceSliceToInt(Permissions)

			for _, permID := range permissions.PermissionIDs {
				permIDStr := fmt.Sprintf("%d", permID)
				if perm, ok := allPermissions[permIDStr]; ok {
					Permission += "|" + perm.Slug
				}
			}
			permissions.Can = strings.TrimPrefix(Permission, "|")
		}

		user.UserPermissionByCompany = append(user.UserPermissionByCompany, permissions)
	}

	return user, nil
}

func (stg userRepo) GetList(companyID string) ([]models.User, error) {
	rows, err := stg.db.Query(`SELECT 
								u.id, 
								u.fullname, 
								u.phone,
								c.id,
								u.is_active
								FROM users u 
								LEFT JOIN companies c ON c.id::text = u.company_id::text 
								WHERE c.id is not null and c.id::text = $1`, companyID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(
			&user.ID,
			&user.Fullname,
			&user.Phone,
			&user.CompanyID,
			&user.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (stg userRepo) ChangePassword(userID int64, entity models.ChangePasswordRequest) error {
	if entity.NewPassword != entity.NewPasswordConfirmation {
		return errors.New("confirmation password is not the same with password")
	}
	password, _ := utils.HashPassword(entity.NewPassword)

	query := `UPDATE "users" SET
	password = :password,
	updated_at = now()
	WHERE
	id = :id`

	params := map[string]interface{}{
		"id":       userID,
		"password": password,
	}

	q, arr := helper.ReplaceQueryParams(query, params)
	_, err := stg.db.Exec(q, arr...)
	if err != nil {
		return err
	}
	return err
}

func (stg *userRepo) GetPermissionByPrimaryKey(ID string) (models.Permission, error) {
	var permission models.Permission
	err := stg.db.QueryRow(`select id, slug, name from permissions where id = $1`, ID).Scan(
		&permission.ID,
		&permission.Slug,
		&permission.Name,
	)
	if err != nil {
		return permission, errors.New("permission topilmadi")
	}
	return permission, nil
}

func (stg *userRepo) GetAllPermissions() (map[string]models.Permission, error) {
	permissions := make(map[string]models.Permission)

	rows, err := stg.db.Query(`select id, slug, name from permissions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var permission models.Permission
		if err := rows.Scan(&permission.ID, &permission.Slug, &permission.Name); err != nil {
			return nil, err
		}
		permissions[permission.ID] = permission
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (stg *userRepo) Edit(entity models.UpdateUserRequest) (rowsAffected int64, err error) {
	query := `UPDATE "users" SET `

	if entity.Fullname != "" {
		query += `fullname = :fullname,`
	}

	if entity.Phone != "" {
		query += `phone = :phone,`
	}

	if entity.IsActive != nil {
		query += `is_active = :is_active,`
	}

	query += `updated_at = now()
			  WHERE id = :id `

	params := map[string]interface{}{
		"id":        entity.ID,
		"fullname":  entity.Fullname,
		"phone":     entity.Phone,
		"is_active": entity.IsActive,
	}

	query, arr := helper.ReplaceQueryParams(query, params)

	result, err := stg.db.Exec(query, arr...)

	if err != nil {
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if len(entity.Permissions) > 0 {
		for _, permission := range entity.Permissions {
			var userCompanyID string
			stg.db.QueryRow(`SELECT company_id FROM user_companies WHERE company_id = $1 AND user_id = $2`,
				permission.CompanyID, entity.ID).Scan(&userCompanyID)

			PermissionIDs := utils.SetArray(utils.IntSliceToInterface(permission.PermissionIDs))

			if len(userCompanyID) > 0 {
				query = `UPDATE "user_companies" SET 
					permission_ids = :permission_ids, 
					is_courier = :is_courier, 
					updated_at = now() 
				WHERE company_id = :company_id AND user_id = :user_id`

				permissionEditParams := map[string]interface{}{
					"permission_ids": PermissionIDs,
					"is_courier":     permission.IsCourier,
					"company_id":     permission.CompanyID,
					"user_id":        entity.ID,
				}

				query, arr := helper.ReplaceQueryParams(query, permissionEditParams)

				_, err := stg.db.Exec(query, arr...)

				if err != nil {
					return 0, err
				}
			} else {
				_, err := stg.db.Exec(`INSERT INTO user_companies(
					permission_ids,
					company_id,
					user_id,
					is_courier
				) VALUES (
					$1,
					$2,
					$3,
					$4
				)`,
					PermissionIDs,
					permission.CompanyID,
					entity.ID,
					permission.IsCourier,
				)

				if err != nil {
					return 0, err
				}
			}
		}
	}

	return rowsAffected, nil
}

func (stg *userRepo) GetCouriesList(companyID string) ([]models.GetCouriesResponse, error) {
	var result []models.GetCouriesResponse

	rows, err := stg.db.Query(`
		SELECT u.id, fullname 
		FROM user_companies uc
		INNER JOIN users u ON uc.user_id::text = u.id::text
		WHERE uc.company_id::text = $1 AND uc.is_courier = true`, companyID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var courier models.GetCouriesResponse
		err := rows.Scan(&courier.ID, &courier.Fullname)
		if err != nil {
			return nil, err
		}
		result = append(result, courier)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
