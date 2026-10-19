package integration

import (
	"time"

	"github.com/depsi4/cargos/pkg/entity"
)

type Product struct {
	entity.Base
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Stock       int64  `json:"stock"`
	MinStock    int64  `json:"min_stock"`
	TargetStock int64  `json:"target_stock"`
	Active      bool   `json:"active"`
}
type ProductInput struct {
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	MinStock    int64  `json:"min_stock"`
	TargetStock int64  `json:"target_stock"`
}
type Movement struct {
	entity.Base
	ProductID string `json:"product_id"`
	Quantity  int64  `json:"quantity"`
	Reason    string `json:"reason"`
	Reference string `json:"reference"`
	Balance   int64  `json:"balance"`
}
type Supplier struct {
	entity.Base
	Name     string `json:"name"`
	Email    string `json:"email"`
	LeadDays int    `json:"lead_days"`
	Active   bool   `json:"active"`
}
type SupplierInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	LeadDays int    `json:"lead_days"`
}
type Order struct {
	entity.Base
	ProductID  string    `json:"product_id"`
	SupplierID string    `json:"supplier_id"`
	Quantity   int64     `json:"quantity"`
	UnitPrice  int64     `json:"unit_price"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	ExpectedAt time.Time `json:"expected_at"`
}
type OrderInput struct {
	ProductID  string `json:"product_id"`
	SupplierID string `json:"supplier_id"`
	Quantity   int64  `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
	Currency   string `json:"currency"`
}
type Demand struct {
	entity.Base
	ProductID  string    `json:"product_id"`
	Quantity   int64     `json:"quantity"`
	OccurredAt time.Time `json:"occurred_at"`
	Note       string    `json:"note"`
}
type DemandInput struct {
	ProductID  string    `json:"product_id"`
	Quantity   int64     `json:"quantity"`
	OccurredAt time.Time `json:"occurred_at"`
	Note       string    `json:"note"`
}
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
type Prediction struct {
	entity.Base
	WindowDays       int              `json:"window_days"`
	Recommendations  []Recommendation `json:"recommendations"`
	NotificationSent bool             `json:"notification_sent"`
}
type Notification struct {
	entity.Base
	EventID string `json:"event_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Read    bool   `json:"read"`
}

type tokenPair struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Type    string `json:"token_type"`
	Expires int    `json:"expires_in"`
}
