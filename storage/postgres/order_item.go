package postgres

import (
	"bw-erp/helper"
	"bw-erp/models"
	"bw-erp/storage/repo"

	"github.com/jmoiron/sqlx"
)

type orderItemRepo struct {
	db *sqlx.DB
}

func NewOrderItemRepo(db *sqlx.DB) repo.OrderItemI {
	return &orderItemRepo{db: db}
}

func (stg orderItemRepo) Create(userID int64, entity models.CreateOrderItemModel) error {
	var id int

	if entity.IsCountable {
		entity.Height = 1
		entity.Width = 1
	}

	err := stg.db.QueryRow(`INSERT INTO order_items(
		order_id,
		type,
		price,
		width,
		height,
		status,
		description,
		is_countable,
		order_item_type_id,
		created_by
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10
	) RETURNING id`,
		entity.OrderID,
		entity.ItemType,
		entity.Price,
		entity.Width,
		entity.Height,
		entity.OrderItemStatus,
		entity.Description,
		entity.IsCountable,
		entity.OrderItemTypeID,
		userID,
	).Scan(&id)

	if err != nil {
		return err
	}

	stg.db.QueryRow(`INSERT INTO status_change_histories(
		historyable_id,
		historyable_type,
		user_id,
		status
	) VALUES (
		$1,
		$2,
		$3,
		$4
	) RETURNING id`,
		id,
		"order_items",
		userID,
		entity.OrderItemStatus,
	).Scan()

	return nil
}

func (stg orderItemRepo) Update(entity models.UpdateOrderItemRequest) (rowsAffected int64, err error) {

	if entity.IsCountable {
		entity.Height = 1
		entity.Width = 1
	}

	query := `UPDATE "order_items" SET `

	if entity.Price != 0 {
		query += `price = :price,`
	}
	if entity.Width != 0 {
		query += `width = :width,`
	}
	if entity.Height != 0 {
		query += `height = :height,`
	}
	if entity.Description != "" {
		query += `description = :description,`
	}

	query += `type = :type, is_countable = :is_countable, updated_at = now()
			  WHERE
					id = :id`

	params := map[string]interface{}{
		"id":           entity.ID,
		"price":        entity.Price,
		"width":        entity.Width,
		"height":       entity.Height,
		"description":  entity.Description,
		"type":         entity.ItemType,
		"is_countable": entity.IsCountable,
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

	return rowsAffected, nil
}

func (stg *orderItemRepo) DeleteByID(ID int) error {
	_, err := stg.db.Exec(`delete from order_items where id = $1`, ID)
	if err != nil {
		return err
	}

	return nil
}

func (stg *orderItemRepo) UpdateStatus(userID int64, entity models.UpdateOrderItemStatusRequest) (rowsAffected int64, err error) {
	query := `UPDATE "order_items" SET `

	query += `status = :status WHERE id = :id`

	params := map[string]interface{}{
		"id":     entity.ID,
		"status": entity.OrderItemStatus,
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

	stg.db.QueryRow(`INSERT INTO status_change_histories(
		historyable_id,
		historyable_type,
		user_id,
		status
	) VALUES (
		$1,
		$2,
		$3,
		$4
	) RETURNING id`,
		entity.ID,
		"order_items",
		userID,
		entity.OrderItemStatus,
	).Scan()

	return rowsAffected, nil
}
