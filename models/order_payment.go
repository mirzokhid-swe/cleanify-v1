package models

import (
	"errors"
	"math"
	"strconv"
)

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

// PaymentLimitErrorFor is the optional company rule on payment size: a payment
// (scope "transaction") or the order's paid total (scope "total") must be at
// least percent% of the order total and at most total + that minimum.
// Returns "" when it passes. Orders without a price are free.
func PaymentLimitErrorFor(amount float64, s PaymentSummary, enabled bool, percent int, scope string) string {
	if !enabled || s.Total <= 0 {
		return ""
	}
	min := math.Round(s.Total * float64(percent) / 100)
	max := s.Total + min
	// JS tomoni bilan bir xil: noma'lum qiymat "total" deb qaraladi.
	checked := s.Paid + amount
	if scope == "transaction" {
		checked = amount
	}
	if checked < min {
		return "Kamida " + spaced(min) + " so'm"
	}
	if checked > max {
		return "Ko'pi bilan " + spaced(max) + " so'm"
	}
	return ""
}

// spaced writes a whole number with spaces between thousands: 115000 -> "115 000".
func spaced(n float64) string {
	digits := strconv.FormatInt(int64(n), 10)
	out := ""
	for i, c := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out += " "
		}
		out += string(c)
	}
	return out
}

type DeleteTransactionRequest struct {
	CompanyID string `form:"company_id" binding:"required"`
}

// OrderPaymentRef: o'chirishdan oldin huquqni tekshirish uchun to'lovning kompaniyasi va kiritgan foydalanuvchi.
type OrderPaymentRef struct {
	CompanyID  string
	ReceiverID int64
}
