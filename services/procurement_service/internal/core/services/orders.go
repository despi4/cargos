package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/depsi4/cargos/procurement_service/internal/core/domain"
	"github.com/depsi4/cargos/procurement_service/internal/core/ports"
)

type Orders struct {
	repository ports.OrderRepository
	suppliers  ports.SupplierRepository
	inventory  ports.Inventory
}

func NewOrders(orders ports.OrderRepository, suppliers ports.SupplierRepository, inventory ports.Inventory) *Orders {
	return &Orders{orders, suppliers, inventory}
}
func (s *Orders) Create(ctx context.Context, owner string, in domain.OrderInput) (*domain.Order, error) {
	in.Currency = strings.ToUpper(in.Currency)
	if !entity.ValidID(in.ProductID) || !entity.ValidID(in.SupplierID) || in.Quantity < 1 || in.Quantity > 1_000_000_000 || in.UnitPrice < 0 || in.UnitPrice > 1_000_000_000 || len(in.Currency) != 3 {
		return nil, fault.Invalid("valid product_id, supplier_id, quantity 1..1000000000, unit_price 0..1000000000 and three-letter currency required")
	}
	for _, c := range in.Currency {
		if c < 'A' || c > 'Z' {
			return nil, fault.Invalid("currency must contain three letters")
		}
	}
	if err := s.inventory.CheckActive(ctx, owner, in.ProductID); err != nil {
		return nil, err
	}
	supplier, err := s.suppliers.Get(ctx, owner, in.SupplierID)
	if err != nil {
		return nil, err
	}
	if !supplier.Active {
		return nil, fault.ErrConflict
	}
	order := &domain.Order{Base: entity.New(), ProductID: in.ProductID, SupplierID: in.SupplierID, Quantity: in.Quantity, UnitPrice: in.UnitPrice, Currency: in.Currency, Status: "draft", ExpectedAt: time.Now().UTC().AddDate(0, 0, supplier.LeadDays)}
	if err = s.repository.Create(ctx, owner, order); err != nil {
		return nil, err
	}
	return order, nil
}

// Persist receiving before any remote side effect. Cancellation cannot race a
// credited receipt; the durable state is resumed by RetryReceipts after a crash.
func (s *Orders) ChangeStatus(ctx context.Context, owner, id, status string) (*domain.Order, error) {
	if status != "ordered" && status != "received" && status != "cancelled" {
		return nil, fault.Invalid("status must be ordered, received or cancelled")
	}
	order, err := s.repository.Update(ctx, owner, id, func(order *domain.Order) error {
		if order.Status == status || (order.Status == "receiving" && status == "received") {
			return nil
		}
		allowed := (order.Status == "draft" && (status == "ordered" || status == "cancelled")) || (order.Status == "ordered" && (status == "received" || status == "cancelled"))
		if !allowed {
			return fault.ErrConflict
		}
		order.Status = status
		if status == "received" {
			order.Status = "receiving"
		}
		order.UpdatedAt = time.Now().UTC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	if status == "received" {
		return s.completeReceipt(ctx, owner, id)
	}
	return order, nil
}
func (s *Orders) completeReceipt(ctx context.Context, owner, id string) (*domain.Order, error) {
	return s.repository.Update(ctx, owner, id, func(order *domain.Order) error {
		if order.Status == "received" {
			return nil
		}
		if order.Status != "receiving" {
			return fault.ErrConflict
		}
		if err := s.inventory.Receive(ctx, owner, order); err != nil {
			return err
		}
		order.Status = "received"
		order.UpdatedAt = time.Now().UTC()
		return nil
	})
}
func (s *Orders) RetryReceipts(ctx context.Context) error {
	pending, err := s.repository.Pending(ctx)
	if err != nil {
		return err
	}
	var failures []error
	for _, v := range pending {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if _, err := s.completeReceipt(ctx, v.OwnerID, v.OrderID); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (s *Orders) List(ctx context.Context, owner string, limit, offset int) ([]domain.Order, error) {
	return s.repository.List(ctx, owner, limit, offset)
}
func (s *Orders) Get(ctx context.Context, owner, id string) (*domain.Order, error) {
	return s.repository.Get(ctx, owner, id)
}
