package postgres

import (
	"bw-erp/helper"
	"bw-erp/models"
	"bw-erp/storage/repo"
	"time"

	"github.com/jmoiron/sqlx"
)

type orderRepo struct {
	db *sqlx.DB
}

func NewOrderRepo(db *sqlx.DB) repo.OrderI {
	return &orderRepo{db: db}
}

func (stg *orderRepo) Create(userID int64, entity models.CreateOrderModel) (id int, err error) {
	var status int8

	if entity.Status == 0 {
		status = 1
	} else {
		status = entity.Status
	}

	tx, err := stg.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	var orderNumber string
	err = tx.QueryRow(`
	  SELECT 'a-' || (
		COALESCE(
		  (SELECT MAX((substring(order_number from '\d+$'))::bigint)
		   FROM orders
		   WHERE company_id = $1), 0
		) + 1
	  )`, entity.CompanyID).Scan(&orderNumber)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(`INSERT INTO orders(
		company_id,
		phone,
		count,
		slug,
		description,
		address,
		status,
		client_id,
		order_number
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING id`,
		entity.CompanyID,
		entity.Phone,
		entity.Count,
		entity.Slug,
		entity.Description,
		entity.Address,
		status,
		entity.ClientID,
		orderNumber,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	// Insert status change history
	_, err = tx.Exec(`INSERT INTO status_change_histories(
		historyable_id,
		historyable_type,
		user_id,
		status
	) VALUES ($1, $2, $3, $4)`,
		id,
		"orders",
		userID,
		status,
	)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (stg *orderRepo) GetList(companyID string, queryParam models.OrdersListRequest) (res models.OrderListResponse, err error) {
	var arr []interface{}
	res = models.OrderListResponse{}
	params := make(map[string]interface{})
	query := `SELECT 
		o.id,
		o.status, 
		o.address,
		o.created_at,
		o.phone,
		o.order_number,
		COALESCE(o.courier_id, null) as courier_id, 
		ROUND(CAST(COALESCE(sum(oi.price*oi.width*oi.height), 0) AS NUMERIC), 2) as price
		FROM "orders" as o 
		left join order_items oi on o.id = oi.order_id
		left join order_statuses os on os.number = o.status and os.company_id = o.company_id`

	filter := " WHERE true"
	group := " group by o.id, o.status, o.address, o.created_at, o.phone, o.courier_id, os.order"
	order := " ORDER BY os.order ASC, o.created_at desc"
	offset := " OFFSET 0"
	limit := " LIMIT 20"

	params["company_id"] = companyID
	filter += " and (o.company_id = :company_id)"

	if len(queryParam.Search) >= 3 {
		params["search"] = queryParam.Search
		filter += "AND (o.order_number ILIKE ('%' || :search || '%') OR o.phone ILIKE ('%' || :search || '%') OR o.address ILIKE ('%' || :search || '%'))"
	}

	if queryParam.Status != 0 {
		params["status"] = queryParam.Status
		filter += " AND (o.status = :status)"
	}

	if queryParam.PaymentStatus != 0 {
		params["payment_status"] = queryParam.PaymentStatus
		filter += " AND (o.payment_status = :payment_status)"
	}

	if queryParam.CourierID != "" {
		params["courier_id"] = queryParam.CourierID
		filter += " AND (o.courier_id = :courier_id)"
	}

	if !queryParam.DateFrom.IsZero() {
		params["date_from"] = queryParam.DateFrom
		filter += " AND (o.created_at >= :date_from::date)"
	}

	if !queryParam.DateTo.IsZero() {
		params["date_to"] = queryParam.DateTo
		filter += " AND (o.created_at <= :date_to::date)"
	}

	if queryParam.Offset > 0 {
		params["offset"] = queryParam.Offset
		offset = " OFFSET :offset"
	}

	// if queryParam.Limit > 0 {
	// 	params["limit"] = queryParam.Limit
	// 	limit = " LIMIT :limit"
	// }

	params["limit"] = 150
	limit = " LIMIT :limit"

	cQ := `SELECT count(1) FROM "orders" as o` + filter
	cQ, arr = helper.ReplaceQueryParams(cQ, params)
	err = stg.db.QueryRow(cQ, arr...).Scan(
		&res.Count,
	)

	if err != nil {
		return res, err
	}

	q := query + filter + group + order + offset + limit

	q, arr = helper.ReplaceQueryParams(q, params)
	rows, err := stg.db.Query(q, arr...)

	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.OrderList
		err = rows.Scan(
			&order.ID,
			&order.Status,
			&order.Address,
			&order.CreatedAt,
			&order.Phone,
			&order.OrderNumber,
			&order.CourierID,
			&order.Price,
		)
		if err != nil {
			return res, err
		}
		res.Data = append(res.Data, order)
	}

	if err = rows.Err(); err != nil {
		return res, err
	}

	return res, nil
}

func (stg *orderRepo) GetByStatus(companyID string, Status int) (order []models.Order, err error) {
	rows, err := stg.db.Query(`select  
		id,
		address, 
		phone
	from orders 
	where company_id = $1 and status = $2`, companyID, Status)
	if err != nil {
		return nil, err
	}

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err = rows.Scan(
			&order.ID,
			&order.Address,
			&order.PhoneNumber,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (stg *orderRepo) GetByPhone(companyID string, Phone string) (models.Order, error) {
	var order models.Order
	err := stg.db.QueryRow(`select id, company_id, phone, count, slug, description, latitute, longitude, created_at, updated_at from orders where company_id = $1 and phone = $2`, companyID, Phone).Scan(
		&order.ID,
		&order.CompanyID,
		&order.PhoneNumber,
		&order.Count,
		&order.Slug,
		&order.Description,
		&order.Latitute,
		&order.Longitude,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return order, err
	}

	return order, nil
}

func (stg *orderRepo) GetDetailedByPrimaryKey(ID int) (models.OrderShowResponse, error) {
	var order models.OrderShowResponse
	err := stg.db.QueryRow(`select o.id, 
									o.company_id, 
									COALESCE(c.phone_number, ''), 
									COALESCE(c.additional_phone_number, ''), 
									COALESCE(c.work_number, ''), 
									o.count, 
									COALESCE(o.slug, ''), 
									COALESCE(o.description, ''), 
									c.latitute, 
									c.longitude, 
									COALESCE(o.client_id, 0), 
									COALESCE(o.address, ''),
									o.status,
									o.uuid,
									COALESCE(o.payment_status, 0),
									coalesce(o.service_price, 0),
									coalesce(o.discount_percentage, 0),
									coalesce(o.discounted_price, 0),
									coalesce(round(sum(oi.width * oi.height * oi.price)::numeric, 2), 0) as price,
									COALESCE(o.courier_id, null) as courier_id,
									o.created_at,
									o.updated_at 
								from orders o
								left join clients c on o.client_id = c.id 
								left join order_items oi on oi.order_id = o.id
								where o.id = $1 group by 
									o.id, 
									o.company_id, 
									c.phone_number, 
									c.additional_phone_number, 
									c.work_number, 
									o.count, 
									o.slug, 
									o.description, 
									c.latitute, 
									c.longitude, 
									o.client_id, 
									o.uuid,
									o.address,
									o.status,
									o.payment_status,
									o.service_price,
									o.discount_percentage,
									o.discounted_price,
									o.created_at,
									o.updated_at `, ID).Scan(
		&order.ID,
		&order.CompanyID,
		&order.PhoneNumber,
		&order.AdditionalPhoneNumber,
		&order.WorkNumber,
		&order.Count,
		&order.Slug,
		&order.Description,
		&order.Latitute,
		&order.Longitude,
		&order.ClientID,
		&order.Address,
		&order.Status,
		&order.Uuid,
		&order.PaymentStatus,
		&order.ServicePrice,
		&order.DiscountPercentage,
		&order.DiscountPrice,
		&order.Price,
		&order.CourierID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return order, err
	}

	rows, err := stg.db.Query(`select id, order_id, type, price, width, height, status, is_countable, description, order_item_type_id from order_items where order_id = $1 order by created_at`, ID)

	if err != nil {
		return order, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.Type, &item.Price, &item.Width, &item.Height, &item.OrderItemStatus, &item.IsCountable, &item.Description, &item.OrderItemTypeID); err != nil {
			return order, err
		}
		data, err := stg.db.Query(`select sch.status, u.fullname, sch.created_at from status_change_histories sch inner join users u on u.id = sch.user_id where historyable_type = 'order_items' and historyable_id = $1 order by created_at desc`, item.ID)
		if err != nil {
			return order, err
		}
		defer data.Close()

		for data.Next() {
			var history models.StatusChangeHistory
			if err := data.Scan(&history.Status, &history.Fullname, &history.CreatedAt); err != nil {
				return order, err
			}
			item.StatusChangeHistory = append(item.StatusChangeHistory, history)
		}
		order.OrderItems = append(order.OrderItems, item)
	}

	rows, err = stg.db.Query(`select sch.status, u.fullname, sch.created_at from status_change_histories sch inner join users u on u.id = sch.user_id where historyable_type = 'orders' and historyable_id = $1 order by created_at asc`, ID)
	if err != nil {
		return order, err
	}
	defer rows.Close()

	for rows.Next() {
		var history models.StatusChangeHistory
		if err := rows.Scan(&history.Status, &history.Fullname, &history.CreatedAt); err != nil {
			return order, err
		}
		order.StatusChangeHistory = append(order.StatusChangeHistory, history)
	}

	transactions, err := stg.db.Query(`select u.fullname, t.payment_type, t.amount, t.created_at from transactions t
									inner join users u on t.receiver_type = 'users' and t.receiver_id = u.id
									where payer_type = 'orders' and payer_id = $1`, order.ID)
	if err != nil {
		return order, err
	}
	defer transactions.Close()

	for transactions.Next() {
		var transaction models.OrderTransaction
		if err := transactions.Scan(&transaction.ReceiverFullname, &transaction.PaymentType, &transaction.Amount, &transaction.CreatedAt); err != nil {
			return order, err
		}
		order.OrderTransaction = append(order.OrderTransaction, transaction)
	}

	comments, err := stg.db.Query(`
		SELECT 
			c.id,
			c.model_type,
			c.model_id,
			c.type,
			c.message,
			c.voice_url,
			COALESCE(u.fullname, '') as full_name,
			c.created_at
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.model_type = 'order' AND c.model_id = $1 
		ORDER BY c.created_at DESC`, ID)
	if err != nil {
		return order, err
	}
	defer comments.Close()

	for comments.Next() {
		var comment models.Comment
		if err := comments.Scan(
			&comment.ID,
			&comment.ModelType,
			&comment.ModelID,
			&comment.Type,
			&comment.Message,
			&comment.VoiceURL,
			&comment.FullName,
			&comment.CreatedAt,
		); err != nil {
			return order, err
		}
		order.Comments = append(order.Comments, comment)
	}

	return order, nil
}

func (stg *orderRepo) GetByPrimaryKey(ID int) (models.OrderShowResponse, error) {
	var order models.OrderShowResponse
	err := stg.db.QueryRow(`select o.id, 
									o.company_id, 
									COALESCE(c.phone_number, ''), 
									COALESCE(c.additional_phone_number, ''), 
									COALESCE(c.work_number, ''), 
									o.count, 
									coalesce(o.slug, ''), 
									o.status,
									coalesce(o.description, ''), 
									c.latitute, 
									c.longitude, 
									COALESCE(o.client_id, 0), 
									COALESCE(o.address, ''),
									ROUND(CAST(COALESCE(sum(oi.price*oi.width*oi.height), 0) AS NUMERIC), 2) as price,
									round(cast(COALESCE(sum(oi.width*oi.height), 0) as numeric), 2) as square, 
									o.courier_id,
									o.created_at,
									o.updated_at 
								from orders o
								left join clients c on o.client_id = c.id 
								left join order_items oi on o.id = oi.order_id
								where o.id = $1 group by o.id, 	o.company_id, c.phone_number, c.additional_phone_number, c.work_number, o.count,o.slug, o.description, 
								c.latitute, c.longitude, o.client_id, o.status, o.address,o.created_at,o.updated_at`, ID).Scan(
		&order.ID,
		&order.CompanyID,
		&order.PhoneNumber,
		&order.AdditionalPhoneNumber,
		&order.WorkNumber,
		&order.Count,
		&order.Slug,
		&order.Status,
		&order.Description,
		&order.Latitute,
		&order.Longitude,
		&order.ClientID,
		&order.Address,
		&order.Price,
		&order.Square,
		&order.CourierID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return order, err
	}

	return order, nil
}

func (stg *orderRepo) GetByUuid(uuid string) (models.OrderReceipt, error) {
	var order models.OrderReceipt
	err := stg.db.QueryRow(`select c.name, o.address, o.phone, o.id, o.service_price, o.discount_percentage, o.discounted_price, o.created_at::date from orders o inner join companies c on o.company_id = c.id where uuid = $1`, uuid).Scan(
		&order.CompanyName,
		&order.Address,
		&order.Phone,
		&order.OrderNumber,
		&order.ServicePrice,
		&order.DiscountPercentage,
		&order.DiscountedPrice,
		&order.CreatedAt,
	)
	if err != nil {
		return order, err
	}

	rows, err := stg.db.Query(`select type, price, width, height, is_countable from order_items where order_id = $1 order by created_at`, order.OrderNumber)

	if err != nil {
		return order, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.Type, &item.Price, &item.Width, &item.Height, &item.IsCountable); err != nil {
			return order, err
		}
		order.OrderItems = append(order.OrderItems, item)
	}

	return order, nil
}

func (stg *orderRepo) Update(userID int64, entity *models.UpdateOrderRequest) (rowsAffected int64, err error) {
	query := `UPDATE "orders" SET `

	if entity.Slug != "" {
		query += `slug = :slug,`
	}

	if entity.Status != 0 {
		query += `status = :status,`
	}

	if entity.Phone != "" {
		query += `phone = :phone,`
	}

	if entity.Count != 0 {
		query += `count = :count,`
	}

	if entity.Description != "" {
		query += `description = :description,`
	}

	if entity.Address != "" {
		query += `address = :address,`
	}

	if entity.Latitute != 0 {
		query += `latitute = :latitute,`
	}

	if entity.Longitude != 0 {
		query += `longitude = :longitude,`
	}

	if entity.DiscountedPrice != 0 {
		query += `discounted_price = :discounted_price,`
	}

	if entity.PaymentStatus != 0 {
		query += `payment_status = :payment_status,`
	}

	if entity.CourierID != 0 {
		query += `courier_id = :courier_id,`
	} else {
		query += `courier_id = NULL,`
	}

	query += `updated_at = now()
              WHERE
                    id = :id`

	order, _ := stg.GetByPrimaryKey(entity.ID)
	if entity.Longitude != 0 && entity.Latitute != 0 && order.ClientID != 0 {
		updateOrderQuery := `UPDATE "clients" SET longitude = $1, latitute = $2, address = $3 WHERE id = $4`
		clientParams := []interface{}{
			entity.Longitude,
			entity.Latitute,
			entity.Address,
			order.ClientID,
		}

		_, err := stg.db.Exec(updateOrderQuery, clientParams...)
		if err != nil {
			return 0, err
		}
	}

	params := map[string]interface{}{
		"id":               entity.ID,
		"status":           entity.Status,
		"slug":             entity.Slug,
		"phone":            entity.Phone,
		"description":      entity.Description,
		"count":            entity.Count,
		"address":          entity.Address,
		"latitute":         entity.Latitute,
		"longitude":        entity.Longitude,
		"payment_status":   entity.PaymentStatus,
		"courier_id":       entity.CourierID,
		"discounted_price": entity.DiscountedPrice,
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

func (stg *orderRepo) GetLocation(ID int) (models.Order, error) {
	var order models.Order
	err := stg.db.QueryRow(`select o.id, 
								   o.company_id, 
								   c.phone_number, 
								   c.additional_phone_number, 
								   c.work_number, 
								   o.count, 
								   o.slug, 
								   o.description, 
								   c.latitute, 
								   c.longitude, 
								   o.created_at, 
								   o.updated_at 
								from orders o
							left join clients c on o.client_id = c.id 
							where o.id = $1`, ID).Scan(
		&order.ID,
		&order.CompanyID,
		&order.PhoneNumber,
		&order.AdditionalPhoneNumber,
		&order.WorkNumber,
		&order.Count,
		&order.Slug,
		&order.Description,
		&order.Latitute,
		&order.Longitude,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return order, err
	}

	return order, nil
}

func (stg *orderRepo) Delete(entity models.DeleteOrderRequest) error {
	_, err := stg.db.Exec(`delete from order_items where order_id = $1`, entity.ID)
	if err != nil {
		return err
	}

	_, err = stg.db.Exec(`delete from orders where id = $1 and company_id = $2`, entity.ID, entity.CompanyID)
	if err != nil {
		return err
	}
	return nil
}

func (stg *orderRepo) SetOrderPrice(entity models.SetOrderPriceRequest) error {
	query := `UPDATE "orders" SET discounted_price = :discounted_price, payment_status = :payment_status where id = :id`

	var payment_status models.PaymentStatus = models.Pending

	params := map[string]interface{}{
		"id":               entity.ID,
		"discounted_price": entity.DiscountedPrice,
		"payment_status":   payment_status,
	}

	query, arr := helper.ReplaceQueryParams(query, params)
	result, err := stg.db.Exec(query, arr...)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (stg *orderRepo) AddComment(entity models.CreateOrderComment) error {
	query := `
		INSERT INTO comments (
			model_type,
			model_id,
			type,
			message,
			voice_url,
			user_id,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := stg.db.Exec(
		query,
		"order",
		entity.OrderID,
		entity.Type,
		entity.Message,
		entity.VoiceURL,
		entity.UserID,
		time.Now(),
	)

	return err
}
