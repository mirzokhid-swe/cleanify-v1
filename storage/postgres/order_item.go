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

// recalcPaymentStatus: an item change moves the total, so the payment status is
// recomputed (a manual status is kept — the function returns without changes).
func (stg orderItemRepo) recalcPaymentStatus(orderID int) {
	stg.db.Exec(`select order_payment_status_recalc($1)`, orderID)
}

func (stg orderItemRepo) orderIDOfItem(itemID int) int {
	var orderID int
	stg.db.QueryRow(`select order_id from order_items where id = $1`, itemID).Scan(&orderID)
	return orderID
}

// GetOrderID returns the order of the item, sql.ErrNoRows when there is no such item.
func (stg orderItemRepo) GetOrderID(itemID int) (int, error) {
	var orderID int
	err := stg.db.QueryRow(`select order_id from order_items where id = $1`, itemID).Scan(&orderID)
	return orderID, err
}

// Create: entity.Rows() ta bir xil buyumni (har biriga holat tarixi bilan) bitta
// tranzaksiyada yozadi; birortasi xato bersa hech narsa qolmaydi.
func (stg orderItemRepo) Create(userID int64, entity models.CreateOrderItemModel) ([]int, error) {
	tx, err := stg.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ids := []int{}
	for i := 0; i < entity.Rows(); i++ {
		var id int
		err := tx.QueryRow(`INSERT INTO order_items(
			order_id, type, price, width, height, status, description,
			is_countable, order_item_type_id, created_by, unit
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
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
			entity.Unit,
		).Scan(&id)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(`INSERT INTO status_change_histories(
			historyable_id, historyable_type, user_id, status
		) VALUES ($1, $2, $3, $4)`,
			id,
			"order_items",
			userID,
			entity.OrderItemStatus,
		)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	stg.recalcPaymentStatus(entity.OrderID)
	return ids, nil
}

// Update: o'lchamni handler MeasureForUnit bilan tayyorlaydi, shuning uchun hamma
// maydon har doim yoziladi (tur almashsa order_item_type_id va unit ham).
func (stg orderItemRepo) Update(entity models.UpdateOrderItemRequest) (rowsAffected int64, err error) {
	result, err := stg.db.Exec(`UPDATE order_items SET
		price = $2,
		width = $3,
		height = $4,
		description = $5,
		type = $6,
		is_countable = $7,
		unit = $8,
		order_item_type_id = $9,
		updated_at = now()
	WHERE id = $1`,
		entity.ID,
		entity.Price,
		entity.Width,
		entity.Height,
		entity.Description,
		entity.ItemType,
		entity.IsCountable,
		entity.Unit,
		entity.OrderItemTypeID,
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return 0, err
	}

	stg.recalcPaymentStatus(stg.orderIDOfItem(entity.ID))
	return rowsAffected, nil
}

func (stg *orderItemRepo) DeleteByID(ID int) error {
	orderID := stg.orderIDOfItem(ID)
	_, err := stg.db.Exec(`delete from order_items where id = $1`, ID)
	if err != nil {
		return err
	}
	stg.recalcPaymentStatus(orderID)
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
