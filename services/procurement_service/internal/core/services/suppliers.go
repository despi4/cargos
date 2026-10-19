package services

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/depsi4/cargos/pkg/entity"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/depsi4/cargos/procurement_service/internal/core/domain"
	"github.com/depsi4/cargos/procurement_service/internal/core/ports"
)

type Suppliers struct{ repository ports.SupplierRepository }

func NewSuppliers(repo ports.SupplierRepository) *Suppliers { return &Suppliers{repository: repo} }
func validateSupplier(in *domain.SupplierInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	addr, err := mail.ParseAddress(in.Email)
	if len(in.Name) < 1 || len(in.Name) > 200 || err != nil || addr.Address != in.Email || len(in.Email) > 255 || in.LeadDays < 0 || in.LeadDays > 365 {
		return fault.Invalid("name, valid email and lead_days 0..365 required")
	}
	return nil
}
func (s *Suppliers) Create(ctx context.Context, owner string, in domain.SupplierInput) (*domain.Supplier, error) {
	if err := validateSupplier(&in); err != nil {
		return nil, err
	}
	v := &domain.Supplier{Base: entity.New(), Name: in.Name, Email: in.Email, LeadDays: in.LeadDays, Active: true}
	if err := s.repository.Create(ctx, owner, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (s *Suppliers) Update(ctx context.Context, owner, id string, in domain.SupplierInput) (*domain.Supplier, error) {
	if err := validateSupplier(&in); err != nil {
		return nil, err
	}
	v, err := s.repository.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	v.Name, v.Email, v.LeadDays = in.Name, in.Email, in.LeadDays
	v.UpdatedAt = time.Now().UTC()
	return s.repository.Update(ctx, owner, v)
}

func (s *Suppliers) List(ctx context.Context, owner string, limit, offset int) ([]domain.Supplier, error) {
	return s.repository.List(ctx, owner, limit, offset)
}
func (s *Suppliers) Get(ctx context.Context, owner, id string) (*domain.Supplier, error) {
	return s.repository.Get(ctx, owner, id)
}

func (s *Suppliers) Archive(ctx context.Context, owner, id string) error {
	return s.repository.Archive(ctx, owner, id)
}
