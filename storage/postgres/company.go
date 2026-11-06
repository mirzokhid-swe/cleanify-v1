package postgres

import (
	"encoding/json"

	"bw-erp/models"
	"bw-erp/storage/repo"

	"github.com/jmoiron/sqlx"
)

type companyRepo struct {
	db *sqlx.DB
}

func NewCompanyRepo(db *sqlx.DB) repo.CompanyStorageI {
	return &companyRepo{db: db}
}

func (stg *companyRepo) Create(id string, entity models.CreateCompanyModel) error {
	_, err := stg.db.Exec(`INSERT INTO companies(
		id,
		name,
		owner_id
	) VALUES (
		$1,
		$2,
		$3
	)`,
		id,
		entity.Name,
		entity.OwnerId,
	)

	if err != nil {
		return err
	}

	return nil
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

func (stg *companyRepo) GetByOwnerId(ownerId int64) ([]models.Company, error) {
	rows, err := stg.db.Query(`select id, name, owner_id from companies where owner_id = $1`, ownerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var company models.Company
		err = rows.Scan(&company.ID, &company.Name)
		if err != nil {
			return nil, err
		}
		companies = append(companies, company)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return companies, nil
}

const companyListSelect = `select c.id, c.name, c.is_blocked,
		(select count(distinct uc.user_id) from user_companies uc where uc.company_id = c.id::text)
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
