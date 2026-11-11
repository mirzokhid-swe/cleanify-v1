package models

import "errors"

type OrderItemStatus int8

const (
	OrderItemStatusDraft    OrderItemStatus = 0
	OrderItemStatusWashed   OrderItemStatus = 1
	OrderItemStatusPrepared OrderItemStatus = 2
)

type CreateOrderItemModel struct {
	OrderID         int     `json:"order_id" binding:"required"`
	OrderItemTypeID string  `json:"order_item_type_id" binding:"required"`
	Price           float32 `json:"price"`
	Width           float32 `json:"width"`
	Height          float32 `json:"height"`
	OrderItemStatus int8    `json:"status"`
	Description     string  `json:"description"`
	// Count: nechta bir xil qator yoziladi ("Soni"), bo'sh = 1.
	// Eng katta qiymat kompaniya sozlamasidan (ItemCountMax), handler tekshiradi.
	Count int `json:"count" binding:"omitempty,gte=1"`
	// Quyidagilarni handler turdan to'ldiradi, mijoz yubora olmaydi.
	ItemType    string `json:"-"`
	IsCountable bool   `json:"-"`
	Unit        string `json:"-"`
}

// Rows: nechta qator yoziladi.
func (m CreateOrderItemModel) Rows() int {
	if m.Count < 1 {
		return 1
	}
	return m.Count
}

type UpdateOrderItemRequest struct {
	ID              int     `json:"id" binding:"required"`
	OrderItemTypeID string  `json:"order_item_type_id" binding:"required"`
	Price           float32 `json:"price"`
	Width           float32 `json:"width"`
	Height          float32 `json:"height"`
	Description     string  `json:"description"`
	// Quyidagilarni handler turdan to'ldiradi, mijoz yubora olmaydi.
	ItemType    string `json:"-"`
	IsCountable bool   `json:"-"`
	Unit        string `json:"-"`
}

var (
	ErrAreaRequired     = errors.New("width and height are required")
	ErrQuantityRequired = errors.New("quantity is required")
)

// MeasureForUnit: so'rovdagi o'lchamni birlikka qarab saqlanadigan width, height ga aylantiradi.
// Narx hamma joyda price × width × height, shuning uchun:
//   - area (m²): bo'yi va eni, ikkalasi > 0;
//   - piece (dona): 1 × 1;
//   - length (metr), weight (kg): width = miqdor, height = 1. Eski front ikkala tomonni
//     yuboradi (3×1 yoki 1×3), shuning uchun miqdor = width × height (0 o'rniga 1).
func MeasureForUnit(unit string, width, height float32) (float32, float32, error) {
	switch unit {
	case "piece":
		return 1, 1, nil
	case "length", "weight":
		if width == 0 && height == 0 {
			return 0, 0, ErrQuantityRequired
		}
		if width == 0 {
			width = 1
		}
		if height == 0 {
			height = 1
		}
		return width * height, 1, nil
	default:
		if width <= 0 || height <= 0 {
			return 0, 0, ErrAreaRequired
		}
		return width, height, nil
	}
}

type UpdateOrderItemStatusRequest struct {
	OrderItemStatus OrderItemStatus `json:"status" binding:"required"`
	ID              int             `json:"id" binding:"required"`
}

type OrderItem struct {
	ID                  int                   `json:"id,omitempty"`
	OrderID             int                   `json:"order_id,omitempty"`
	Type                string                `json:"type,omitempty"`
	Price               float32               `json:"price,omitempty"`
	Width               float32               `json:"width,omitempty"`
	Height              float32               `json:"height,omitempty"`
	OrderItemStatus     OrderItemStatus       `json:"status,omitempty"`
	IsCountable         bool                  `json:"is_countable,omitempty"`
	Unit                string                `json:"unit,omitempty"`
	Description         string                `json:"description,omitempty"`
	OrderItemTypeID     *string               `json:"order_item_type_id,omitempty"`
	StatusChangeHistory []StatusChangeHistory `json:"status_change_history,omitempty"`
}
