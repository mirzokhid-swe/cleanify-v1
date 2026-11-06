package postgres

import (
	"encoding/json"

	"bw-erp/helper"
	"bw-erp/models"
)

const ordersListMaxLimit = 150

// ordersListFilter builds the WHERE of the base CTE: company, search (3+ chars:
// order_number, phone, address), period (date_to inclusive). Pure.
func ordersListFilter(companyID string, q models.OrdersListRequest) (string, map[string]interface{}) {
	params := map[string]interface{}{"company_id": companyID}
	filter := " WHERE o.company_id = :company_id"
	if len(q.Search) >= 3 {
		params["search"] = "%" + q.Search + "%"
		filter += " AND (o.order_number ILIKE :search OR o.phone ILIKE :search OR o.address ILIKE :search)"
	}
	if !q.DateFrom.IsZero() {
		params["date_from"] = q.DateFrom
		filter += " AND o.created_at >= :date_from::date"
	}
	if !q.DateTo.IsZero() {
		params["date_to"] = q.DateTo
		filter += " AND o.created_at < (:date_to::date + 1)"
	}
	return filter, params
}

// ordersListDataFilter narrows the list (not the counts) by status / debtor /
// legacy payment_status, courier_id. Appended after "WHERE true". Pure.
func ordersListDataFilter(q models.OrdersListRequest) string {
	f := ""
	if q.Status != 0 {
		f += " AND status = :status"
	}
	if q.Debtor {
		f += " AND status = 6 AND payment_status IN (1,2)"
	}
	if q.PaymentStatus != 0 {
		f += " AND payment_status = :payment_status"
	}
	if q.CourierID != "" {
		f += " AND courier_id = :courier_id"
	}
	return f
}

func ordersListLimit(limit int32) int32 {
	if limit <= 0 || limit > ordersListMaxLimit {
		return ordersListMaxLimit
	}
	return limit
}

// base: one row per order with price (items) and paid (non-deleted payments).
const ordersListBaseSQL = `WITH base AS (
	SELECT o.id, o.order_number, o.phone, o.address, o.description, o.status, o.payment_status,
	       o.courier_id, cu.fullname AS courier_name, o.created_at, os."order" AS status_order,
	       p.price, t.paid
	FROM orders o
	LEFT JOIN users cu ON cu.id = o.courier_id
	LEFT JOIN order_statuses os ON os.company_id = o.company_id AND os.number = o.status
	LEFT JOIN LATERAL (
		SELECT ROUND(COALESCE(SUM(oi.price * oi.width * oi.height), 0)::numeric, 2) AS price
		FROM order_items oi WHERE oi.order_id = o.id
	) p ON true
	LEFT JOIN LATERAL (
		SELECT COALESCE(SUM(tr.amount), 0) AS paid
		FROM transactions tr WHERE tr.payer_type = 'orders' AND tr.payer_id = o.id AND tr.deleted_at IS NULL
	) t ON true`

func (stg *orderRepo) GetList(companyID string, q models.OrdersListRequest) (models.OrderListResponse, error) {
	res := models.OrderListResponse{Data: []models.OrderList{}, Counts: models.OrderListCounts{Statuses: map[string]int{}}}
	if !q.DateFrom.IsZero() && !q.DateTo.IsZero() && q.DateFrom.After(q.DateTo) {
		return res, models.ErrBadDateRange
	}

	filter, params := ordersListFilter(companyID, q)
	base := ordersListBaseSQL + filter + `
)`
	dataFilter := ordersListDataFilter(q)
	if q.Status != 0 {
		params["status"] = q.Status
	}
	if q.PaymentStatus != 0 {
		params["payment_status"] = q.PaymentStatus
	}
	if q.CourierID != "" {
		params["courier_id"] = q.CourierID
	}
	params["limit"] = ordersListLimit(q.Limit)
	params["offset"] = q.Offset
	if q.Offset < 0 {
		params["offset"] = 0
	}

	countsSQL, args := helper.ReplaceQueryParams(base+`
SELECT count(*),
       count(*) FILTER (WHERE status = 6 AND payment_status IN (1,2)),
       COALESCE((SELECT jsonb_object_agg(status, n) FROM (SELECT status, count(*) AS n FROM base GROUP BY status) s), '{}'::jsonb)
FROM base`, params)
	var statusesJSON []byte
	if err := stg.db.QueryRow(countsSQL, args...).Scan(&res.Counts.All, &res.Counts.Debtor, &statusesJSON); err != nil {
		return res, err
	}
	if len(statusesJSON) > 0 {
		if err := json.Unmarshal(statusesJSON, &res.Counts.Statuses); err != nil {
			return res, err
		}
	}

	totalSQL, args := helper.ReplaceQueryParams(base+`
SELECT count(*) FROM base WHERE true`+dataFilter, params)
	if err := stg.db.QueryRow(totalSQL, args...).Scan(&res.Count); err != nil {
		return res, err
	}

	dataSQL, args := helper.ReplaceQueryParams(base+`
SELECT id, order_number, phone, address, description, status, payment_status, courier_id, courier_name,
       created_at, price, paid, GREATEST(price - paid, 0) AS debt
FROM base WHERE true`+dataFilter+`
ORDER BY status_order ASC NULLS LAST, created_at DESC
LIMIT :limit OFFSET :offset`, params)
	rows, err := stg.db.Query(dataSQL, args...)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var o models.OrderList
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.Phone, &o.Address, &o.Description, &o.Status, &o.PaymentStatus,
			&o.CourierID, &o.CourierName, &o.CreatedAt, &o.Price, &o.Paid, &o.Debt); err != nil {
			return res, err
		}
		res.Data = append(res.Data, o)
	}
	return res, rows.Err()
}
