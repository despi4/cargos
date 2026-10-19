package domain

import (
	"github.com/depsi4/cargos/pkg/entity"
)

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
