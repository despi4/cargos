package http

import (
	"net/http"

	"github.com/depsi4/cargos/internal/platform/httpx"
	"github.com/depsi4/cargos/notification_service/internal/core/services"
)

func Mount(mux *http.ServeMux, s *services.Notifications) {
	mux.HandleFunc("GET /api/v1/notifications", httpx.List(s.List))
	mux.HandleFunc("GET /api/v1/notifications/{id}", httpx.Get(s.Get))
	mux.HandleFunc("PATCH /api/v1/notifications/{id}", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Read *bool `json:"read"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		if in.Read == nil {
			return httpx.Bad("read is required")
		}
		v, err := s.SetRead(r.Context(), httpx.Owner(r), r.PathValue("id"), *in.Read)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, v)
		return nil
	}))
	mux.HandleFunc("POST /internal/notifications", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			EventID string `json:"event_id"`
			Title   string `json:"title"`
			Message string `json:"message"`
		}
		if err := httpx.Decode(w, r, &in); err != nil {
			return err
		}
		v, err := s.Publish(r.Context(), httpx.Owner(r), in.EventID, in.Title, in.Message)
		if err != nil {
			return err
		}
		if v == nil {
			w.WriteHeader(204)
		} else {
			httpx.JSON(w, 201, v)
		}
		return nil
	}))
}
