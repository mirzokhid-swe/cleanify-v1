package models

import "testing"

func TestPaymentAmountErrorForItemlessOrder(t *testing.T) {
	open := PaymentSummary{Total: 0, Paid: 0, Remaining: 0}
	if PaymentAmountErrorFor(5000, open) != "" {
		t.Fatal("itemless order accepts any positive amount")
	}
	if PaymentAmountErrorFor(0, open) != "Summa noto'g'ri" {
		t.Fatal("zero still rejected")
	}
	priced := PaymentSummary{Total: 100, Paid: 0, Remaining: 100}
	if PaymentAmountErrorFor(101, priced) != "Qolgan qarzdan ko'p" {
		t.Fatal("priced order keeps the remaining check")
	}
}

func TestPaymentAmountError(t *testing.T) {
	if PaymentAmountError(100, 100) != "" {
		t.Fatal("equal is ok")
	}
	if PaymentAmountError(50, 100) != "" {
		t.Fatal("less is ok")
	}
	if PaymentAmountError(101, 100) != "Qolgan qarzdan ko'p" {
		t.Fatal("over")
	}
	if PaymentAmountError(0, 100) != "Summa noto'g'ri" {
		t.Fatal("zero")
	}
	if PaymentAmountError(-5, 100) != "Summa noto'g'ri" {
		t.Fatal("negative")
	}
	if PaymentAmountError(1, 0) != "Qolgan qarzdan ko'p" {
		t.Fatal("nothing remaining")
	}
}

func TestPaymentLimitErrorFor(t *testing.T) {
	priced := PaymentSummary{Total: 100000, Paid: 0, Remaining: 100000}
	paid := PaymentSummary{Total: 100000, Paid: 10000, Remaining: 90000}
	itemless := PaymentSummary{Total: 0}
	tests := []struct {
		name    string
		amount  float64
		s       PaymentSummary
		enabled bool
		percent int
		scope   string
		want    string
	}{
		{"disabled", 1, priced, false, 15, "total", ""},
		{"itemless free", 1, itemless, true, 15, "total", ""},
		{"transaction below min", 14999, priced, true, 15, "transaction", "Kamida 15 000 so'm"},
		{"transaction at min", 15000, priced, true, 15, "transaction", ""},
		{"transaction at max", 115000, priced, true, 15, "transaction", ""},
		{"transaction above max", 115001, priced, true, 15, "transaction", "Ko'pi bilan 115 000 so'm"},
		{"total below min", 4999, paid, true, 15, "total", "Kamida 15 000 so'm"},
		{"total at min", 5000, paid, true, 15, "total", ""},
		{"total above max", 105001, paid, true, 15, "total", "Ko'pi bilan 115 000 so'm"},
		{"total paid already enough", 1, PaymentSummary{Total: 100000, Paid: 20000}, true, 15, "total", ""},
	}
	for _, tc := range tests {
		if got := PaymentLimitErrorFor(tc.amount, tc.s, tc.enabled, tc.percent, tc.scope); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
