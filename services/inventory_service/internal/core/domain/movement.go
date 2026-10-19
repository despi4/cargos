package domain

import (
	"github.com/depsi4/cargos/pkg/entity"
)

type Movement struct {
	entity.Base
	ProductID string `json:"product_id"`
	Quantity  int64  `json:"quantity"`
	Reason    string `json:"reason"`
	Reference string `json:"reference"`
	Balance   int64  `json:"balance"`
}
