package postgres

import (
	"bw-erp/helper"
	"bw-erp/models"
	"bw-erp/storage/repo"

	"github.com/jmoiron/sqlx"
)

type orderItemTypeRepo struct {
	db *sqlx.DB
}

func NewOrderItemTypeRepo(db *sqlx.DB) repo.OrderItemTypeI {
	return &orderItemTypeRepo{db: db}
}

const orderItemTypeSelect = `select
		o.id,
		o.name,
		o.price,
		o.unit,
		o.is_countable,
		c.name,
		c.id
	from order_item_types o
	inner join companies c on c.id = o.company_id`

func (stg *orderItemTypeRepo) Create(id string, entity models.OrderItemTypeModel) error {
	_, err := stg.db.Exec(`INSERT INTO order_item_types(
		id,
		name,
		company_id,
		price,
		unit,
		is_countable
	) VALUES ($1, $2, $3, $4, $5, $6)`,
		id,
		entity.Name,
		entity.CompanyID,
		entity.Price,
		entity.Unit,
		models.UnitIsCountable(entity.Unit),
	)
	return err
}

func (stg *orderItemTypeRepo) GetByCompany(CompanyID string) ([]models.OrderItemByCompany, error) {
	rows, err := stg.db.Query(orderItemTypeSelect+`
		where o.company_id = $1
		order by `+models.OrderItemUnitOrderSQL+`, o.price`, CompanyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderItemTypes := []models.OrderItemByCompany{}
	for rows.Next() {
		var orderItemType models.OrderItemByCompany
		err = rows.Scan(
			&orderItemType.ID,
			&orderItemType.Name,
			&orderItemType.Price,
			&orderItemType.Unit,
			&orderItemType.IsCountable,
			&orderItemType.CopmanyName,
			&orderItemType.CopmanyID,
		)
		if err != nil {
			return nil, err
		}
		orderItemTypes = append(orderItemTypes, orderItemType)
	}

	return orderItemTypes, rows.Err()
}

func (stg *orderItemTypeRepo) GetById(orderItemTypeId string, companyID string) (models.OrderItemByCompany, error) {
	var orderItemType models.OrderItemByCompany
	query := orderItemTypeSelect + ` where o.id = $1`
	args := []interface{}{orderItemTypeId}
	if companyID != "" {
		query += ` and o.company_id = $2`
		args = append(args, companyID)
	}
	err := stg.db.QueryRow(query, args...).Scan(
		&orderItemType.ID,
		&orderItemType.Name,
		&orderItemType.Price,
		&orderItemType.Unit,
		&orderItemType.IsCountable,
		&orderItemType.CopmanyName,
		&orderItemType.CopmanyID,
	)
	if err != nil {
		return orderItemType, err
	}

	return orderItemType, nil
}

func (stg *orderItemTypeRepo) Update(entity models.EditOrderItemTypeRequest) (rowsAffected int64, err error) {
	query := `UPDATE "order_item_types"
		SET price = :price, unit = :unit, is_countable = :is_countable, name = :name, updated_at = now()
		WHERE id = :id and company_id = :company_id`

	params := map[string]interface{}{
		"id":           entity.ID,
		"price":        entity.Price,
		"company_id":   entity.CopmanyID,
		"unit":         entity.Unit,
		"is_countable": models.UnitIsCountable(entity.Unit),
		"name":         entity.Name,
	}

	query, arr := helper.ReplaceQueryParams(query, params)
	result, err := stg.db.Exec(query, arr...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (stg *orderItemTypeRepo) Delete(id string, companyID string) (rowsAffected int64, err error) {
	result, err := stg.db.Exec(`delete from order_item_types where id = $1 and company_id = $2`, id, companyID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
