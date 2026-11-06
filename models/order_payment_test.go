package models

import "testing"

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
