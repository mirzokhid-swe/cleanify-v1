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
