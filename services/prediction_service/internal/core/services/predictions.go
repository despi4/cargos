package services

import (
	"context"
	"time"

	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/depsi4/cargos/prediction_service/internal/core/domain"
	"github.com/depsi4/cargos/prediction_service/internal/core/ports"
)

type Predictions struct {
	repository ports.Repository
	source     ports.Source
	notifier   ports.Notifier
}

func New(repo ports.Repository, source ports.Source, notifier ports.Notifier) *Predictions {
	return &Predictions{repo, source, notifier}
}
func (s *Predictions) Create(ctx context.Context, owner string, days int, supplierID string) (*domain.Prediction, error) {
	if days == 0 {
		days = 30
	}
	if days < 1 || days > 365 {
		return nil, fault.Invalid("window_days must be 1..365")
	}
	if supplierID != "" && !entity.ValidID(supplierID) {
		return nil, fault.Invalid("invalid supplier_id")
	}
	products, err := s.source.Products(ctx, owner)
	if err != nil {
		return nil, err
	}
	demands, err := s.source.Demands(ctx, owner)
	if err != nil {
		return nil, err
	}
	suppliers, err := s.source.Suppliers(ctx, owner)
	if err != nil {
		return nil, err
	}
	orders, err := s.source.Orders(ctx, owner)
	if err != nil {
		return nil, err
	}
	var supplier *domain.SupplierSnapshot
	for i := range suppliers {
		v := &suppliers[i]
		if !v.Active {
			continue
		}
		if supplierID != "" {
			if v.ID == supplierID {
				supplier = v
				break
			}
		} else if supplier == nil || v.LeadDays < supplier.LeadDays {
			supplier = v
		}
	}
	if supplierID != "" && supplier == nil {
		return nil, fault.Invalid("supplier not found or archived")
	}
	v := &domain.Prediction{Base: entity.New(), WindowDays: days, Recommendations: Calculate(products, demands, orders, supplier, days, time.Now())}
	v.NotificationSent = len(v.Recommendations) == 0
	if err = s.repository.Create(ctx, owner, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Predictions) List(ctx context.Context, owner string, limit, offset int) ([]domain.Prediction, error) {
	return s.repository.List(ctx, owner, limit, offset)
}

func (s *Predictions) Get(ctx context.Context, owner, id string) (*domain.Prediction, error) {
	return s.repository.Get(ctx, owner, id)
}
