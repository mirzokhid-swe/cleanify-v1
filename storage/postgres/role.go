package postgres

import (
	"bw-erp/models"
	"bw-erp/storage/repo"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type roleRepo struct {
	db *sqlx.DB
}

func NewRoleRepo(db *sqlx.DB) repo.RoleI {
	return &roleRepo{db: db}
}

const roleSelect = `select r.id, r.company_id, r.name, r.permission_ids,
	(select count(*) from user_companies uc join users u on u.id = uc.user_id where uc.role_id = r.id and u.deleted_at is null) as users_count
	from roles r`

func scanRole(scan func(dest ...any) error) (models.Role, error) {
	var role models.Role
	var ids pq.Int64Array
	if err := scan(&role.ID, &role.CompanyID, &role.Name, &ids, &role.UsersCount); err != nil {
		return role, err
	}
	role.PermissionIDs = make([]int, 0, len(ids))
	for _, id := range ids {
		role.PermissionIDs = append(role.PermissionIDs, int(id))
	}
	return role, nil
}

func permissionArray(ids []int) pq.Int64Array {
	arr := make(pq.Int64Array, 0, len(ids))
	for _, id := range ids {
		arr = append(arr, int64(id))
	}
	return arr
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func (stg *roleRepo) GetList(companyID string) ([]models.Role, error) {
	rows, err := stg.db.Query(roleSelect+` where r.company_id = $1 order by r.name`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	roles := []models.Role{}
	for rows.Next() {
		role, err := scanRole(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (stg *roleRepo) GetByID(id int64, companyID string) (models.Role, error) {
	row := stg.db.QueryRow(roleSelect+` where r.id = $1 and r.company_id = $2`, id, companyID)
	return scanRole(row.Scan)
}

func (stg *roleRepo) Create(in models.RoleInput) (int64, error) {
	var id int64
	err := stg.db.QueryRow(
		`insert into roles (company_id, name, permission_ids) values ($1, $2, $3) returning id`,
		in.CompanyID, in.Name, permissionArray(in.PermissionIDs),
	).Scan(&id)
	if isUniqueViolation(err) {
		return 0, models.ErrRoleNameTaken
	}
	if err != nil {
		return 0, fmt.Errorf("create role: %w", err)
	}
	return id, nil
}

func (stg *roleRepo) Update(id int64, in models.RoleInput) (int64, error) {
	res, err := stg.db.Exec(
		`update roles set name = $1, permission_ids = $2, updated_at = now() where id = $3 and company_id = $4`,
		in.Name, permissionArray(in.PermissionIDs), id, in.CompanyID,
	)
	if isUniqueViolation(err) {
		return 0, models.ErrRoleNameTaken
	}
	if err != nil {
		return 0, fmt.Errorf("update role: %w", err)
	}
	return res.RowsAffected()
}

func (stg *roleRepo) Delete(id int64, companyID string) (int64, error) {
	role, err := stg.GetByID(id, companyID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if role.UsersCount > 0 {
		return 0, models.ErrRoleInUse
	}
	res, err := stg.db.Exec(`delete from roles where id = $1 and company_id = $2`, id, companyID)
	if err != nil {
		return 0, fmt.Errorf("delete role: %w", err)
	}
	return res.RowsAffected()
}
