package domain

import "time"

// Snapshots are the prediction service's input contracts, not other services' entities.
type ProductSnapshot struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Stock       int64  `json:"stock"`
	MinStock    int64  `json:"min_stock"`
	TargetStock int64  `json:"target_stock"`
	Active      bool   `json:"active"`
}
type DemandSnapshot struct {
	ProductID  string    `json:"product_id"`
	Quantity   int64     `json:"quantity"`
	OccurredAt time.Time `json:"occurred_at"`
}
type SupplierSnapshot struct {
	ID       string `json:"id"`
	LeadDays int    `json:"lead_days"`
	Active   bool   `json:"active"`
}
type OrderSnapshot struct {
	ProductID string `json:"product_id"`
	Quantity  int64  `json:"quantity"`
	Status    string `json:"status"`
}
type PendingDelivery struct {
	OwnerID      string
	PredictionID string
	Count        int
}
