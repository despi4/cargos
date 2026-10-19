package http

import (
	"net/http"

	"github.com/depsi4/cargos/internal/platform/httpx"
	"github.com/depsi4/cargos/procurement_service/internal/core/domain"
	"github.com/depsi4/cargos/procurement_service/internal/core/services"
)

func Mount(mux *http.ServeMux, suppliers *services.Suppliers, orders *services.Orders) {
	mux.HandleFunc("GET /api/v1/suppliers", httpx.List(suppliers.List))
	mux.HandleFunc("GET /api/v1/suppliers/{id}", httpx.Get(suppliers.Get))
	mux.HandleFunc("POST /api/v1/suppliers", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in domain.SupplierInput
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := suppliers.Create(r.Context(), httpx.Owner(r), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, v)
		return nil
	}))
	mux.HandleFunc("PUT /api/v1/suppliers/{id}", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in domain.SupplierInput
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := suppliers.Update(r.Context(), httpx.Owner(r), r.PathValue("id"), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, v)
		return nil
	}))
	mux.HandleFunc("DELETE /api/v1/suppliers/{id}", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		if err := suppliers.Archive(r.Context(), httpx.Owner(r), r.PathValue("id")); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	mux.HandleFunc("GET /api/v1/orders", httpx.List(orders.List))
	mux.HandleFunc("GET /api/v1/orders/{id}", httpx.Get(orders.Get))
	mux.HandleFunc("POST /api/v1/orders", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in domain.OrderInput
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := orders.Create(r.Context(), httpx.Owner(r), in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, v)
		return nil
	}))
	mux.HandleFunc("POST /api/v1/orders/{id}/status", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Status string `json:"status"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := orders.ChangeStatus(r.Context(), httpx.Owner(r), r.PathValue("id"), in.Status)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, v)
		return nil
	}))
}
