package postgres

import (
	"encoding/json"

	"bw-erp/models"
	"bw-erp/storage/repo"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type companyRepo struct {
	db *sqlx.DB
}

func NewCompanyRepo(db *sqlx.DB) repo.CompanyStorageI {
	return &companyRepo{db: db}
}

// Create: korxona va uning standartlari (models/company_defaults.go) bitta
// tranzaksiyada yoziladi; birortasi xato bersa hech narsa qolmaydi.
func (stg *companyRepo) Create(id string, name string, userID int64) error {
	tx, err := stg.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`insert into companies (id, name) values ($1, $2)`, id, name)
	if isUniqueViolation(err) {
		return models.ErrCompanyNameTaken
	}
	if err != nil {
		return err
	}

	for _, status := range models.DefaultOrderStatuses {
		_, err = tx.Exec(`insert into order_statuses (company_id, name, color, number, "order", slug, description)
			values ($1, $2, $3, $4, $5, $6, $7)`,
			id, status.Name, status.Color, status.Number, status.Order, status.Slug, status.Description)
		if err != nil {
			return err
		}
	}

	for _, itemType := range models.DefaultItemTypes {
		_, err = tx.Exec(`insert into order_item_types (id, company_id, name, price, unit, is_countable)
			values ($1, $2, $3, $4, $5, $6)`,
			uuid.New().String(), id, itemType.Name, itemType.Price, itemType.Unit, models.UnitIsCountable(itemType.Unit))
		if err != nil {
			return err
		}
	}

	var adminRoleID int64
	for _, role := range models.DefaultRoles {
		var roleID int64
		if len(role.Slugs) == 0 {
			err = tx.QueryRow(`insert into roles (company_id, name, permission_ids)
				select $1::uuid, $2::varchar, coalesce(array_agg(id order by id), '{}')
				from permissions where scope = 'company' returning id`,
				id, role.Name).Scan(&roleID)
			adminRoleID = roleID
		} else {
			err = tx.QueryRow(`insert into roles (company_id, name, permission_ids)
				select $1::uuid, $2::varchar, coalesce(array_agg(id order by id), '{}')
				from permissions where slug = any($3) returning id`,
				id, role.Name, pq.Array(role.Slugs)).Scan(&roleID)
		}
		if err != nil {
			return err
		}
	}

	for _, purpose := range models.DefaultPaymentPurposes {
		_, err = tx.Exec(`insert into payment_purposes (company_id, name, slug, type, created_at, updated_at)
			values ($1, $2, $3, $4, now(), now())`,
			id, purpose.Name, purpose.Slug, purpose.Type)
		if err != nil {
			return err
		}
	}

	// Yaratgan super admin Admin roli bilan bog'lanadi: busiz u korxonaga kira olmaydi.
	_, err = tx.Exec(`insert into user_companies (user_id, company_id, role_id, permission_ids, is_courier, created_at, updated_at)
		values ($1, $2, $3, '{}', false, now(), now())`,
		userID, id, adminRoleID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (stg *companyRepo) GetById(id string) (models.Company, error) {
	var company models.Company
	err := stg.db.QueryRow(`select id, name from companies where id = $1`, id).Scan(
		&company.ID,
		&company.Name,
	)
	if err != nil {
		return company, err
	}

	return company, nil
}

const companyListSelect = `select c.id, c.name, c.is_blocked,
		(select count(distinct uc.user_id) from user_companies uc
			join users u on u.id = uc.user_id
			where uc.company_id = c.id::text and u.deleted_at is null)
	from companies c`

func (stg *companyRepo) GetList() ([]models.CompanyListItem, error) {
	rows, err := stg.db.Query(companyListSelect + ` order by c.is_blocked, c.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.CompanyListItem{}
	for rows.Next() {
		var item models.CompanyListItem
		if err := rows.Scan(&item.ID, &item.Name, &item.IsBlocked, &item.UsersCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (stg *companyRepo) GetShow(id string) (models.CompanyShow, error) {
	var show models.CompanyShow
	err := stg.db.QueryRow(companyListSelect+` where c.id = $1`, id).Scan(
		&show.ID, &show.Name, &show.IsBlocked, &show.UsersCount,
	)
	if err != nil {
		return show, err
	}

	rows, err := stg.db.Query(`select key, value from company_settings where company_id = $1`, id)
	if err != nil {
		return show, err
	}
	defer rows.Close()

	stored := map[string]json.RawMessage{}
	for rows.Next() {
		var key string
		var value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return show, err
		}
		stored[key] = json.RawMessage(value)
	}
	if err := rows.Err(); err != nil {
		return show, err
	}
	show.Settings = models.MergeSettings(stored)
	return show, nil
}

func (stg *companyRepo) UpsertSetting(companyID string, key string, value json.RawMessage, userID int64) error {
	_, err := stg.db.Exec(`insert into company_settings (company_id, key, value, updated_by)
		values ($1, $2, $3::jsonb, $4)
		on conflict (company_id, key) do update
		set value = excluded.value, updated_at = now(), updated_by = excluded.updated_by`,
		companyID, key, string(value), userID)
	return err
}

func (stg *companyRepo) SetBlocked(companyID string, blocked bool, userID int64) error {
	_, err := stg.db.Exec(`update companies
		set is_blocked = $2,
		    blocked_at = case when $2 then now() else null end,
		    blocked_by = case when $2 then $3::bigint else null end
		where id = $1`, companyID, blocked, userID)
	return err
}
