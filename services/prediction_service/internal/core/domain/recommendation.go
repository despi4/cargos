package domain

type Recommendation struct {
	ProductID         string  `json:"product_id"`
	ProductName       string  `json:"product_name"`
	Stock             int64   `json:"stock"`
	DailyDemand       float64 `json:"daily_demand"`
	ReorderPoint      int64   `json:"reorder_point"`
	SuggestedQuantity int64   `json:"suggested_quantity"`
	SupplierID        string  `json:"supplier_id,omitempty"`
	LeadDays          int     `json:"lead_days"`
}
