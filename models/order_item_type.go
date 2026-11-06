package models

// OrderItemUnits lists the measurement units of an order item type in the
// order the list is shown: area (m²), piece (dona), length (m), weight (kg).
var OrderItemUnits = []string{"area", "piece", "length", "weight"}

// UnitIsCountable reports whether items of this unit are counted (piece)
// rather than measured. It keeps the legacy is_countable column in sync.
func UnitIsCountable(unit string) bool {
	return unit == "piece"
}

// OrderItemUnitOrderSQL orders rows of alias "o" by unit in OrderItemUnits order.
const OrderItemUnitOrderSQL = "array_position(array['area','piece','length','weight']::text[], o.unit)"

type OrderItemTypeModel struct {
	Name      string  `json:"name" binding:"required,min=2,max=255"`
	Price     float32 `json:"price" binding:"required,gt=0"`
	CompanyID string  `json:"company_id" binding:"required"`
	Unit      string  `json:"unit" binding:"required,oneof=area piece length weight"`
}

type OrderItemByCompany struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
	Unit        string  `json:"unit"`
	IsCountable bool    `json:"is_countable"`
	CopmanyName string  `json:"company_name"`
	CopmanyID   string  `json:"company_id"`
}

type EditOrderItemTypeRequest struct {
	CopmanyID string  `json:"company_id" binding:"required"`
	ID        string  `json:"id" binding:"required"`
	Price     float32 `json:"price" binding:"required,gt=0"`
	Unit      string  `json:"unit" binding:"required,oneof=area piece length weight"`
	Name      string  `json:"name" binding:"required,min=2,max=255"`
}

type GetOrderItemTypeByCompany struct {
	CompanyID string `json:"company_id" form:"company_id" binding:"required"`
}

type DeleteOrderItemTypeRequest struct {
	CompanyID string `form:"company_id" binding:"required"`
}
