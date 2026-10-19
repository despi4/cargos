package http

import (
	"net/http"

	"github.com/depsi4/cargos/internal/platform/httpx"
	"github.com/depsi4/cargos/prediction_service/internal/core/services"
)

func Mount(mux *http.ServeMux, s *services.Predictions) {
	mux.HandleFunc("GET /api/v1/predictions", httpx.List(s.List))
	mux.HandleFunc("GET /api/v1/predictions/{id}", httpx.Get(s.Get))
	mux.HandleFunc("POST /api/v1/predictions", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			WindowDays int    `json:"window_days"`
			SupplierID string `json:"supplier_id"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := s.Create(r.Context(), httpx.Owner(r), in.WindowDays, in.SupplierID)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, v)
		return nil
	}))
}
