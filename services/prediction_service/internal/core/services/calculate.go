package services

import (
	"math"
	"time"

	"github.com/depsi4/cargos/prediction_service/internal/core/domain"
)

const maxQuantity int64 = 1_000_000_000

// Calculate applies the replenishment rules without network or database access.
func Calculate(products []domain.ProductSnapshot, demands []domain.DemandSnapshot, orders []domain.OrderSnapshot, supplier *domain.SupplierSnapshot, days int, now time.Time) []domain.Recommendation {
	totals := map[string]float64{}
	cutoff := now.AddDate(0, 0, -days)
	for _, d := range demands {
		if !d.OccurredAt.Before(cutoff) && !d.OccurredAt.After(now) {
			totals[d.ProductID] += float64(d.Quantity)
		}
	}
	incoming := map[string]int64{}
	for _, o := range orders {
		if o.Status == "ordered" {
			incoming[o.ProductID] += o.Quantity
		}
	}
	lead, supplierID := 7, ""
	if supplier != nil {
		lead, supplierID = supplier.LeadDays, supplier.ID
	}
	out := []domain.Recommendation{}
	for _, p := range products {
		if !p.Active {
			continue
		}
		daily := totals[p.ID] / float64(days)
		reorder := int64(math.Min(float64(maxQuantity), math.Max(float64(p.MinStock), math.Ceil(daily*float64(lead)))))
		target := int64(math.Min(float64(maxQuantity), math.Max(float64(p.TargetStock), math.Ceil(daily*float64(lead+7)))))
		available := p.Stock + incoming[p.ID]
		quantity := target - available
		if available <= reorder && quantity > 0 {
			out = append(out, domain.Recommendation{ProductID: p.ID, ProductName: p.Name, Stock: p.Stock, DailyDemand: daily, ReorderPoint: reorder, SuggestedQuantity: quantity, SupplierID: supplierID, LeadDays: lead})
		}
	}
	return out
}
