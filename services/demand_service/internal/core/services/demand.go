package services

import (
	"context"
	"time"

	"github.com/depsi4/cargos/demand_service/internal/core/domain"
	"github.com/depsi4/cargos/demand_service/internal/core/ports"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
)

type Demand struct {
	records   ports.Repository
	inventory ports.Inventory
}

func New(records ports.Repository, inventory ports.Inventory) *Demand {
	return &Demand{records, inventory}
}
func (s *Demand) Create(ctx context.Context, owner string, in domain.DemandInput) (*domain.Demand, error) {
	if in.OccurredAt.IsZero() {
		in.OccurredAt = time.Now().UTC()
	}
	if !entity.ValidID(in.ProductID) || in.Quantity < 1 || in.Quantity > 1_000_000_000 || in.OccurredAt.After(time.Now().Add(time.Minute)) || len(in.Note) > 500 {
		return nil, fault.Invalid("valid product_id, quantity 1..1000000000, non-future occurred_at and note <=500 bytes required")
	}
	if err := s.inventory.CheckActive(ctx, owner, in.ProductID); err != nil {
		return nil, err
	}
	v := &domain.Demand{Base: entity.New(), ProductID: in.ProductID, Quantity: in.Quantity, OccurredAt: in.OccurredAt, Note: in.Note}
	if err := s.records.Create(ctx, owner, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Demand) List(ctx context.Context, owner string, limit, offset int) ([]domain.Demand, error) {
	return s.records.List(ctx, owner, limit, offset)
}

func (s *Demand) Get(ctx context.Context, owner, id string) (*domain.Demand, error) {
	return s.records.Get(ctx, owner, id)
}
