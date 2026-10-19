package http

import (
	"net/http"
	"strings"

	"github.com/depsi4/cargos/auth_service/internal/core/ports"
	"github.com/depsi4/cargos/internal/platform/httpx"
)

type Handler struct{ service ports.AuthService }
type credentials struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Mount(mux *http.ServeMux, s ports.AuthService) {
	h := &Handler{service: s}
	mux.HandleFunc("POST /api/v1/auth/register", httpx.Handle(h.register))
	mux.HandleFunc("POST /api/v1/auth/login", httpx.Handle(h.login))
	mux.HandleFunc("POST /api/v1/auth/refresh", httpx.Handle(h.refresh))
	mux.HandleFunc("POST /api/v1/auth/logout", httpx.Handle(h.logout))
	mux.HandleFunc("GET /api/v1/auth/me", httpx.Handle(h.me))
}
func (h *Handler) register(w http.ResponseWriter, r *http.Request) error {
	var in credentials
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	v, err := h.service.Register(r.Context(), in.Username, in.Email, in.Password)
	if err != nil {
		return err
	}
	httpx.JSON(w, 201, v)
	return nil
}
func (h *Handler) login(w http.ResponseWriter, r *http.Request) error {
	var in credentials
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	v, err := h.service.Login(r.Context(), in.Email, in.Password, r.RemoteAddr, r.UserAgent())
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "no-store")
	httpx.JSON(w, 200, v)
	return nil
}
func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Refresh string `json:"refresh_token"`
	}
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	v, err := h.service.RefreshToken(r.Context(), in.Refresh, r.RemoteAddr, r.UserAgent())
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "no-store")
	httpx.JSON(w, 200, v)
	return nil
}
func bearer(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}
func (h *Handler) me(w http.ResponseWriter, r *http.Request) error {
	v, err := h.service.CurrentUser(r.Context(), bearer(r))
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "no-store")
	httpx.JSON(w, 200, v)
	return nil
}
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) error {
	if err := h.service.LogoutAccess(r.Context(), bearer(r)); err != nil {
		return err
	}
	w.WriteHeader(204)
	return nil
}
