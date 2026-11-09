package postgres

import (
	"bw-erp/helper"
	"bw-erp/models"
	"bw-erp/pkg/utils"
	"bw-erp/storage/repo"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) repo.UserI {
	return &userRepo{db: db}
}

func (stg userRepo) Create(entity models.CreateUserModel) (int64, error) {
	tx, err := stg.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRow(`INSERT INTO users(
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
	if isPhoneTaken(err) {
		return 0, models.ErrPhoneTaken
	}
	if err != nil {
		return 0, err
	}

	for _, in := range entity.Permissions {
		if err := writeUserCompany(tx, id, in); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// isPhoneTaken: faol xodimlar orasida raqam takrorlandi (users_phone_active_key).
func isPhoneTaken(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "users_phone_active_key"
}

// writeUserCompany: RoleID berilsa rol shu korxonaniki ekani tekshiriladi va faqat
// role_id, is_courier yoziladi (permission_ids saqlanib qoladi); aks holda role_id = NULL
// va o'z permission_ids yoziladi.
func writeUserCompany(tx *sqlx.Tx, userID int64, in models.UserCompanyInput) error {
	if in.RoleID != nil {
		var roleCompanyID string
		err := tx.QueryRow(`select company_id::text from roles where id = $1`, *in.RoleID).Scan(&roleCompanyID)
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrRoleForeign
		}
		if err != nil {
			return err
		}
		if roleCompanyID != in.CompanyID {
			return models.ErrRoleForeign
		}
		_, err = tx.Exec(`insert into user_companies (user_id, company_id, role_id, permission_ids, is_courier, created_at, updated_at)
			values ($1, $2, $3, '{}', $4, now(), now())
			on conflict (user_id, company_id) do update set role_id = excluded.role_id, is_courier = excluded.is_courier, updated_at = now()`,
			userID, in.CompanyID, *in.RoleID, in.IsCourier)
		return err
	}

	_, err := tx.Exec(`insert into user_companies (user_id, company_id, role_id, permission_ids, is_courier, created_at, updated_at)
		values ($1, $2, null, $3, $4, now(), now())
		on conflict (user_id, company_id) do update set role_id = null, permission_ids = excluded.permission_ids, is_courier = excluded.is_courier, updated_at = now()`,
		userID, in.CompanyID, permissionArray(in.PermissionIDs), in.IsCourier)
	return err
}

func (stg userRepo) GetByPhone(phone string) (models.AuthUserModel, error) {
	var user models.AuthUserModel
	err := stg.db.QueryRow(`SELECT id, phone, password from users where phone = $1 and deleted_at is null`, phone).Scan(
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
	err := stg.db.QueryRow(`select u.id, u.fullname, u.phone, c.id, u.is_active, u.is_super_admin from users u left join companies c on c.id::text = u.company_id::text where u.id = $1 and u.deleted_at is null`, id).Scan(
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

	rows, err := stg.db.Query(`select c.id, c.name, uc.permission_ids, uc.role_id, r.name, r.permission_ids, uc.is_courier, c.is_blocked,
								coalesce((select jsonb_object_agg(cs.key, cs.value) from company_settings cs where cs.company_id = c.id), '{}'::jsonb)
								from user_companies uc
								inner join companies c on c.id::text = uc.company_id::text
								left join roles r on r.id = uc.role_id
								where uc.user_id = $1 order by c.name desc`, user.ID)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	for rows.Next() {
		var permissions models.UserPermissionByCompany
		var ownIDs, roleIDs pq.Int64Array
		var settings []byte
		if err := rows.Scan(&permissions.CompanyID, &permissions.CompanyName, &ownIDs, &permissions.RoleID, &permissions.RoleName, &roleIDs, &permissions.IsCourier, &permissions.IsBlocked, &settings); err != nil {
			return user, err
		}
		permissions.Settings = json.RawMessage(settings)

		permissions.PermissionIDs = make([]int, 0, len(ownIDs))
		for _, permID := range ownIDs {
			permissions.PermissionIDs = append(permissions.PermissionIDs, int(permID))
		}

		// rol biriktirilgan bo'lsa huquqlar roldan, aks holda o'z permission_ids dan
		canIDs := ownIDs
		if permissions.RoleID != nil {
			canIDs = roleIDs
		}
		Permission := ""
		for _, permID := range canIDs {
			permIDStr := fmt.Sprintf("%d", permID)
			if perm, ok := allPermissions[permIDStr]; ok {
				Permission += "|" + perm.Slug
			}
		}
		permissions.Can = strings.TrimPrefix(Permission, "|")

		user.UserPermissionByCompany = append(user.UserPermissionByCompany, permissions)
	}
	if err = rows.Err(); err != nil {
		return user, err
	}

	return user, nil
}

func (stg userRepo) GetList(companyID string) ([]models.UserListItem, error) {
	rows, err := stg.db.Query(`select u.id, coalesce(u.fullname, ''), coalesce(u.phone, ''), coalesce(u.is_active, false),
								coalesce(uc.is_courier, false), uc.role_id, r.name, uc.permission_ids
								from user_companies uc
								join users u on u.id = uc.user_id
								left join roles r on r.id = uc.role_id
								where uc.company_id::text = $1 and u.deleted_at is null
								order by u.fullname`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.UserListItem{}
	for rows.Next() {
		var user models.UserListItem
		var ids pq.Int64Array
		err = rows.Scan(&user.ID, &user.Fullname, &user.Phone, &user.IsActive, &user.IsCourier, &user.RoleID, &user.RoleName, &ids)
		if err != nil {
			return nil, err
		}
		user.PermissionIDs = make([]int, 0, len(ids))
		for _, id := range ids {
			user.PermissionIDs = append(user.PermissionIDs, int(id))
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
	userID, err := strconv.ParseInt(entity.ID, 10, 64)
	if err != nil {
		return 0, err
	}

	tx, err := stg.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

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
		"id":        userID,
		"fullname":  entity.Fullname,
		"phone":     entity.Phone,
		"is_active": entity.IsActive,
	}

	query, arr := helper.ReplaceQueryParams(query, params)

	result, err := tx.Exec(query, arr...)
	if isPhoneTaken(err) {
		return 0, models.ErrPhoneTaken
	}
	if err != nil {
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return 0, err
	}

	for _, in := range entity.Permissions {
		if err := writeUserCompany(tx, userID, in); err != nil {
			return 0, err
		}
	}

	// Filial dan olib tashlash: chaqiruvchi boshqaradigan, lekin permissions[] da yo'q korxonalar
	sent := []string{}
	for _, in := range entity.Permissions {
		sent = append(sent, in.CompanyID)
	}
	_, err = tx.Exec(`delete from user_companies
		where user_id = $1 and company_id::text = any($2::text[]) and not (company_id::text = any($3::text[]))`,
		userID, pq.Array(entity.ManageableCompanyIDs), pq.Array(sent))
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (stg *userRepo) GetCouriesList(companyID string) ([]models.GetCouriesResponse, error) {
	var result []models.GetCouriesResponse

	rows, err := stg.db.Query(`
		SELECT u.id, fullname 
		FROM user_companies uc
		INNER JOIN users u ON uc.user_id::text = u.id::text
		WHERE uc.company_id::text = $1 AND uc.is_courier = true AND u.deleted_at IS NULL`, companyID)

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

// Delete: soft delete. role_id tozalanadi (FK rolni o'chirishga to'sqinlik qilmasin),
// permission_ids o'zgarmaydi.
func (stg *userRepo) Delete(id int64, deletedBy int64) (int64, error) {
	tx, err := stg.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`update users set deleted_at = now(), deleted_by = $2, is_active = false, chat_id = null, updated_at = now()
		where id = $1 and deleted_at is null`, id, deletedBy)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if _, err := tx.Exec(`update user_companies set role_id = null where user_id = $1`, id); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}
