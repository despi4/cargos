package domain

import (
	"github.com/depsi4/cargos/pkg/entity"
	"time"
)

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
