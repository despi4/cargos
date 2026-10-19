package services

import (
	"testing"
	"time"

	"github.com/depsi4/cargos/prediction_service/internal/core/domain"
)

func TestRecommendations(t *testing.T) {
	now := time.Now()
	p := domain.ProductSnapshot{ID: "product-1", Name: "Tyre", Stock: 2, MinStock: 10, TargetStock: 50, Active: true}
	supplier := &domain.SupplierSnapshot{ID: "supplier-1", LeadDays: 7}
	history := []domain.DemandSnapshot{{ProductID: p.ID, Quantity: 90, OccurredAt: now.Add(-time.Hour)}, {ProductID: p.ID, Quantity: 900, OccurredAt: now.AddDate(0, 0, -31)}, {ProductID: p.ID, Quantity: 900, OccurredAt: now.Add(time.Hour)}}
	got := Calculate([]domain.ProductSnapshot{p}, history, nil, supplier, 30, now)
	if len(got) != 1 || got[0].DailyDemand != 3 || got[0].ReorderPoint != 21 || got[0].SuggestedQuantity != 48 {
		t.Fatalf("unexpected recommendation: %+v", got)
	}
	orders := []domain.OrderSnapshot{{ProductID: p.ID, Quantity: 48, Status: "ordered"}}
	if r := Calculate([]domain.ProductSnapshot{p}, history, orders, supplier, 30, now); len(r) != 0 {
		t.Fatalf("incoming stock ignored: %+v", r)
	}
	orders[0].Status = "cancelled"
	if len(Calculate([]domain.ProductSnapshot{p}, history, orders, supplier, 30, now)) != 1 {
		t.Fatal("cancelled order incorrectly counts as stock")
	}
	p.Active = false
	if len(Calculate([]domain.ProductSnapshot{p}, history, nil, supplier, 30, now)) != 0 {
		t.Fatal("archived product recommended")
	}
}
