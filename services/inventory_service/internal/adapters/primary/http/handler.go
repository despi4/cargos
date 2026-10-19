package http

import (
	"net/http"

	"github.com/depsi4/cargos/internal/platform/httpx"
	"github.com/depsi4/cargos/inventory_service/internal/core/domain"
	"github.com/depsi4/cargos/inventory_service/internal/core/services"
)

type Handler struct{ service *services.Inventory }

func Mount(mux *http.ServeMux, s *services.Inventory) {
	h := &Handler{s}
	mux.HandleFunc("GET /api/v1/products", httpx.List(s.ListProducts))
	mux.HandleFunc("GET /api/v1/products/{id}", httpx.Get(s.GetProduct))
	mux.HandleFunc("GET /api/v1/movements", httpx.List(s.ListMovements))
	mux.HandleFunc("POST /api/v1/products", httpx.Handle(h.create))
	mux.HandleFunc("PUT /api/v1/products/{id}", httpx.Handle(h.update))
	mux.HandleFunc("DELETE /api/v1/products/{id}", httpx.Handle(h.archive))
	mux.HandleFunc("POST /api/v1/products/{id}/movements", httpx.Handle(h.move))
	mux.HandleFunc("POST /internal/products/{id}/receipts", httpx.Handle(h.receive))
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var in domain.ProductInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	v, err := h.service.Create(r.Context(), httpx.Owner(r), in)
	if err != nil {
		return err
	}
	httpx.JSON(w, 201, v)
	return nil
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) error {
	var in domain.ProductInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	v, err := h.service.Update(r.Context(), httpx.Owner(r), r.PathValue("id"), in)
	if err != nil {
		return err
	}
	httpx.JSON(w, 200, v)
	return nil
}
func (h *Handler) archive(w http.ResponseWriter, r *http.Request) error {
	if err := h.service.ArchiveProduct(r.Context(), httpx.Owner(r), r.PathValue("id")); err != nil {
		return err
	}
	w.WriteHeader(204)
	return nil
}

type movementRequest struct {
	Quantity  int64  `json:"quantity"`
	Reason    string `json:"reason"`
	Reference string `json:"reference"`
}

func (h *Handler) move(w http.ResponseWriter, r *http.Request) error    { return h.adjust(w, r, false) }
func (h *Handler) receive(w http.ResponseWriter, r *http.Request) error { return h.adjust(w, r, true) }
func (h *Handler) adjust(w http.ResponseWriter, r *http.Request, receipt bool) error {
	var in movementRequest
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	v, err := h.service.Adjust(r.Context(), httpx.Owner(r), r.PathValue("id"), in.Quantity, in.Reason, in.Reference, receipt)
	if err != nil {
		return err
	}
	httpx.JSON(w, 200, v)
	return nil
}
