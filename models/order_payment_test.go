package models

import "testing"

func TestPaymentAmountError(t *testing.T) {
	if PaymentAmountError(100) != "" {
		t.Fatal("positive is ok")
	}
	if PaymentAmountError(0) != "Summa noto'g'ri" {
		t.Fatal("zero")
	}
	if PaymentAmountError(-5) != "Summa noto'g'ri" {
		t.Fatal("negative")
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
