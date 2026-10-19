package domain

import (
	"github.com/depsi4/cargos/pkg/entity"
	"time"
)

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

type PendingReceipt struct {
	OwnerID string
	OrderID string
}
