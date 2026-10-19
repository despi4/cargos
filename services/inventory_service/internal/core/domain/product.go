package domain

import (
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
