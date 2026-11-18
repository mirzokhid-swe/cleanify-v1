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

type StatusMetricsRequest struct {
	CompanyID string `form:"company_id" binding:"required"`
	Status    int    `form:"status" binding:"required"`
}

type StatusMetricsItem struct {
	Type string  `json:"type"`
	Unit string  `json:"unit"`
	Qty  float64 `json:"qty"`
}

type StatusMetricsUser struct {
	UserID   int64               `json:"user_id"`
	Fullname string              `json:"fullname"`
	Count    int                 `json:"count"`
	Items    []StatusMetricsItem `json:"items"`
}

type StatusMetricsDay struct {
	Date  string              `json:"date"`
	Count int                 `json:"count"`
	Items []StatusMetricsItem `json:"items"`
	Users []StatusMetricsUser `json:"users"`
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
