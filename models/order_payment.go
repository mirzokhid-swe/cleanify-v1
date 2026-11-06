package models

import "errors"

// PaymentSummary is returned after every payment change and by the order page.
type PaymentSummary struct {
	PaymentStatus *int16  `json:"payment_status"`
	Total         float64 `json:"total"`
	Paid          float64 `json:"paid"`
	Remaining     float64 `json:"remaining"`
}

// ErrPaymentAmount wraps a PaymentAmountError message (handler answers 400).
var ErrPaymentAmount = errors.New("payment amount")

type PaymentAmountErr struct{ Msg string }

func (e PaymentAmountErr) Error() string { return e.Msg }
func (e PaymentAmountErr) Is(target error) bool {
	return target == ErrPaymentAmount
}

// PaymentAmountError returns the user-facing reason a payment amount is
// rejected, or "" when it is acceptable: amount must be > 0 and ≤ remaining.
func PaymentAmountError(amount, remaining float64) string {
	if amount <= 0 {
		return "Summa noto'g'ri"
	}
	if amount > remaining {
		return "Qolgan qarzdan ko'p"
	}
	return ""
}

// PaymentAmountErrorFor applies PaymentAmountError to an order summary. An
// order without items (Total == 0) has no known price yet, so any positive
// amount is accepted — otherwise a courier could never record such a payment.
func PaymentAmountErrorFor(amount float64, s PaymentSummary) string {
	if s.Total == 0 {
		if amount <= 0 {
			return "Summa noto'g'ri"
		}
		return ""
	}
	return PaymentAmountError(amount, s.Remaining)
}

type DeleteTransactionRequest struct {
	CompanyID string `form:"company_id" binding:"required"`
}
