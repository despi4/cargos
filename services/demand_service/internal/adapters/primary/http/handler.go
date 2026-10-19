package http

import (
	"net/http"

	"github.com/depsi4/cargos/demand_service/internal/core/domain"
	"github.com/depsi4/cargos/demand_service/internal/core/services"
	"github.com/depsi4/cargos/internal/platform/httpx"
)

func Mount(mux *http.ServeMux, s *services.Demand) {
	mux.HandleFunc("GET /api/v1/demand", httpx.List(s.List))
	mux.HandleFunc("GET /api/v1/demand/{id}", httpx.Get(s.Get))
	mux.HandleFunc("POST /api/v1/demand", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in domain.DemandInput
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := s.Create(r.Context(), httpx.Owner(r), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, v)
		return nil
	}))
}
