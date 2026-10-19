package services

import (
	"context"
	"strings"
	"time"

	"github.com/depsi4/cargos/inventory_service/internal/core/domain"
	"github.com/depsi4/cargos/inventory_service/internal/core/ports"
	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
)

type Inventory struct {
	products  ports.ProductRepository
	movements ports.MovementRepository
}

func New(products ports.ProductRepository, movements ports.MovementRepository) *Inventory {
	return &Inventory{products: products, movements: movements}
}
func validate(in *domain.ProductInput) error {
	in.SKU = strings.TrimSpace(in.SKU)
	in.Name = strings.TrimSpace(in.Name)
	in.Unit = strings.TrimSpace(in.Unit)
	if len(in.SKU) < 1 || len(in.SKU) > 64 || len(in.Name) < 1 || len(in.Name) > 200 || len(in.Unit) < 1 || len(in.Unit) > 32 {
		return fault.Invalid("sku, name and unit are required (max 64, 200, 32 bytes)")
	}
	if in.MinStock < 0 || in.TargetStock < in.MinStock || in.TargetStock > domain.MaxQuantity {
		return fault.Invalid("require 0 <= min_stock <= target_stock <= 1000000000")
	}
	return nil
}
func (s *Inventory) Create(ctx context.Context, owner string, in domain.ProductInput) (*domain.Product, error) {
	if err := validate(&in); err != nil {
		return nil, err
	}
	p := &domain.Product{Base: entity.New(), SKU: in.SKU, Name: in.Name, Unit: in.Unit, MinStock: in.MinStock, TargetStock: in.TargetStock, Active: true}
	if err := s.products.Create(ctx, owner, p); err != nil {
		return nil, err
	}
	return p, nil
}
func (s *Inventory) Update(ctx context.Context, owner, id string, in domain.ProductInput) (*domain.Product, error) {
	if err := validate(&in); err != nil {
		return nil, err
	}
	p, err := s.products.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	if p.SKU != in.SKU {
		return nil, fault.Invalid("sku is immutable")
	}
	p.Name, p.Unit, p.MinStock, p.TargetStock = in.Name, in.Unit, in.MinStock, in.TargetStock
	p.UpdatedAt = time.Now().UTC()
	return s.products.Update(ctx, owner, p)
}
func (s *Inventory) Adjust(ctx context.Context, owner, id string, quantity int64, reason, reference string, receipt bool) (*domain.Movement, error) {
	reason = strings.TrimSpace(reason)
	reference = strings.TrimSpace(reference)
	if quantity == 0 || quantity > domain.MaxQuantity || quantity < -domain.MaxQuantity || len(reason) < 1 || len(reason) > 500 || len(reference) < 1 || len(reference) > 200 {
		return nil, fault.Invalid("nonzero quantity within +/-1000000000, reason and unique reference required")
	}
	if receipt && quantity < 0 {
		return nil, fault.Invalid("receipt quantity must be positive")
	}
	m := domain.Movement{Base: entity.New(), ProductID: id, Quantity: quantity, Reason: reason, Reference: reference}
	return s.products.Adjust(ctx, owner, id, m, receipt)
}

func (s *Inventory) ListProducts(ctx context.Context, owner string, limit, offset int) ([]domain.Product, error) {
	return s.products.List(ctx, owner, limit, offset)
}

func (s *Inventory) GetProduct(ctx context.Context, owner, id string) (*domain.Product, error) {
	return s.products.Get(ctx, owner, id)
}

func (s *Inventory) ArchiveProduct(ctx context.Context, owner, id string) error {
	return s.products.Archive(ctx, owner, id)
}

func (s *Inventory) ListMovements(ctx context.Context, owner string, limit, offset int) ([]domain.Movement, error) {
	return s.movements.List(ctx, owner, limit, offset)
}
