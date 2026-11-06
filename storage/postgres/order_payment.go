package postgres

import (
	"database/sql"
	"errors"

	"bw-erp/models"

	"github.com/jmoiron/sqlx"
)

// Totals are rounded to whole so'm (item sizes come through float32, so raw
// sums look like 175599.9946 while people pay 175 600).
const paymentSummarySQL = `select o.payment_status,
		round(coalesce((select sum(oi.price * oi.width * oi.height) from order_items oi where oi.order_id = o.id), 0)::numeric),
		round(coalesce((select sum(t.amount) from transactions t where t.payer_type = 'orders' and t.payer_id = o.id and t.deleted_at is null), 0)::numeric)
	from orders o where o.id = $1`

func scanSummary(row *sql.Row) (models.PaymentSummary, error) {
	var s models.PaymentSummary
	var status sql.NullInt16
	if err := row.Scan(&status, &s.Total, &s.Paid); err != nil {
		return s, err
	}
	if status.Valid {
		v := status.Int16
		s.PaymentStatus = &v
	}
	s.Remaining = s.Total - s.Paid
	if s.Remaining < 0 {
		s.Remaining = 0
	}
	return s, nil
}

func (stg *orderRepo) GetPaymentSummary(orderID int) (models.PaymentSummary, error) {
	return scanSummary(stg.db.QueryRow(paymentSummarySQL, orderID))
}

// RecalcPaymentStatus runs order_payment_status_recalc for the order (no-op
// while payment_status_manual is set).
func (stg *orderRepo) RecalcPaymentStatus(orderID int) error {
	_, err := stg.db.Exec(`select order_payment_status_recalc($1)`, orderID)
	return err
}

// recalcTx: money changed, so the manual flag is cleared before recalculating.
func recalcAfterMoneyChange(tx *sqlx.Tx, orderID int) error {
	if _, err := tx.Exec(`update orders set payment_status_manual = false where id = $1`, orderID); err != nil {
		return err
	}
	_, err := tx.Exec(`select order_payment_status_recalc($1)`, orderID)
	return err
}

type queryRower interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

func paymentSummary(q queryRower, orderID int) (models.PaymentSummary, error) {
	return scanSummary(q.QueryRow(paymentSummarySQL, orderID))
}

// AddPayment runs inside one transaction with the order row locked, so two
// simultaneous payments (a double tap) are serialised: the second one sees
// the first and is dropped by the duplicate rule or the remaining check.
func (stg *orderRepo) AddPayment(userID int64, entity models.AddOrderPaymentRequest) (models.PaymentSummary, error) {
	tx, err := stg.db.Beginx()
	if err != nil {
		return models.PaymentSummary{}, err
	}
	defer tx.Rollback()

	var orderID int
	err = tx.QueryRow(`select id from orders where id = $1 and company_id = $2 for update`, entity.OrderID, entity.CompanyID).Scan(&orderID)
	if err != nil {
		return models.PaymentSummary{}, err // sql.ErrNoRows → 404
	}

	// Duplicate within 1 minute by the same user, same order, same amount is ignored.
	var exists bool
	err = tx.QueryRow(`select exists(
			select 1 from transactions
			where company_id = $1 and payer_type = 'orders' and payer_id = $2
			  and receiver_type = 'users' and receiver_id = $3 and amount = $4
			  and deleted_at is null and created_at >= now() - interval '1 minute')`,
		entity.CompanyID, entity.OrderID, userID, entity.Amount).Scan(&exists)
	if err != nil {
		return models.PaymentSummary{}, err
	}
	if exists {
		return paymentSummary(tx, entity.OrderID)
	}

	summary, err := paymentSummary(tx, entity.OrderID)
	if err != nil {
		return summary, err
	}
	if msg := models.PaymentAmountErrorFor(entity.Amount, summary); msg != "" {
		return summary, models.PaymentAmountErr{Msg: msg}
	}

	var paymentPurposeID int
	err = tx.QueryRow(`select id from payment_purposes where slug = 'orders' and company_id = $1`, entity.CompanyID).Scan(&paymentPurposeID)
	if err != nil {
		return summary, err
	}

	_, err = tx.Exec(`insert into transactions(
			company_id, payer_id, payer_type, amount, receiver_id, receiver_type,
			payment_type, payment_purpose_id, description
		) values ($1, $2, 'orders', $3, $4, 'users', $5, $6, $7)`,
		entity.CompanyID, entity.OrderID, entity.Amount, userID, entity.PaymentType, paymentPurposeID, entity.Description)
	if err != nil {
		return summary, err
	}
	if err := recalcAfterMoneyChange(tx, entity.OrderID); err != nil {
		return summary, err
	}
	if err := tx.Commit(); err != nil {
		return summary, err
	}
	return stg.GetPaymentSummary(entity.OrderID)
}

// DeleteTransaction soft-deletes an order payment of the company; the row stays
// (deleted_at, deleted_by) and no longer counts. Missing/other company/already
// deleted/not an order payment → sql.ErrNoRows.
func (stg *orderRepo) DeleteTransaction(userID int64, transactionID int, companyID string) (models.PaymentSummary, error) {
	tx, err := stg.db.Beginx()
	if err != nil {
		return models.PaymentSummary{}, err
	}
	defer tx.Rollback()

	var orderID int
	err = tx.QueryRow(`update transactions set deleted_at = now(), deleted_by = $1
		where id = $2 and company_id = $3 and payer_type = 'orders' and deleted_at is null
		returning payer_id`, userID, transactionID, companyID).Scan(&orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.PaymentSummary{}, sql.ErrNoRows
	}
	if err != nil {
		return models.PaymentSummary{}, err
	}
	if err := recalcAfterMoneyChange(tx, orderID); err != nil {
		return models.PaymentSummary{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.PaymentSummary{}, err
	}
	return stg.GetPaymentSummary(orderID)
}
