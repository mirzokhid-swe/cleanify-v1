package models

import "time"

type WorkVolume struct {
	MeterSquare float64   `json:"meter_square"`
	WashedAt    time.Time `json:"washed_at"`
	Type        string    `json:"type"`
	IsCountable bool      `json:"is_countable"`
}

type WorkVolumeListRequest struct {
	WashedAtFrom time.Time `json:"status,omitempty"`
	WashedAtTo   time.Time `json:"slug,omitempty"`
}

type GetWorkVolumeListRequest struct {
	CompanyID string `json:"company_id" form:"company_id" binding:"required"`
}

type GetServicePaymentStatisticsRequest struct {
	CompanyID string `json:"company_id" form:"company_id" binding:"required"`
	DateFrom  string `json:"date_from" form:"date_from"`
	DateTo    string `json:"date_to" form:"date_to"`
}

type ServicePaymentStatistics struct {
	UserID   string  `json:"user_id"`
	Fullname string  `json:"fullname"`
	Amount   float64 `json:"amount"`
}
