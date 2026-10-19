package httpx

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/depsi4/cargos/pkg/fault"
	"github.com/google/uuid"
)

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string   { return e.Message }
func Bad(message string) error      { return &APIError{400, message} }
func Conflict(message string) error { return &APIError{409, message} }
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func Error(w http.ResponseWriter, err error) {
	code, msg := 500, "internal server error"
	var api *APIError
	var invalid *fault.InvalidInput
	var unavailable *fault.Unavailable
	switch {
	case errors.As(err, &invalid):
		code, msg = 400, invalid.Message
	case errors.As(err, &unavailable):
		code, msg = 503, unavailable.Message
	case errors.Is(err, fault.ErrUnauthorized):
		code, msg = 401, err.Error()
	case errors.As(err, &api):
		code, msg = api.Status, api.Message
	case errors.Is(err, fault.ErrNotFound):
		code, msg = 404, err.Error()
	case errors.Is(err, fault.ErrConflict):
		code, msg = 409, err.Error()
	default:
		slog.Error("request failed", "error", err)
	}
	JSON(w, code, map[string]string{"error": msg})
}
func Decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return Bad("invalid JSON body: " + err.Error())
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return Bad("body must contain one JSON object")
	}
	return nil
}
func Handle(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			Error(w, err)
		}
	}
}
func Owner(r *http.Request) string { return r.Header.Get("X-User-ID") }
func Page(r *http.Request) (int, int, error) {
	limit, offset := 100, 0
	var err error
	if s := r.URL.Query().Get("limit"); s != "" {
		limit, err = strconv.Atoi(s)
		if err != nil {
			return 0, 0, Bad("invalid limit")
		}
	}
	if s := r.URL.Query().Get("offset"); s != "" {
		offset, err = strconv.Atoi(s)
		if err != nil {
			return 0, 0, Bad("invalid offset")
		}
	}
	if limit < 1 || limit > 500 || offset < 0 {
		return 0, 0, Bad("limit must be 1..500 and offset nonnegative")
	}
	return limit, offset, nil
}
func Internal(key string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Internal-Key")), []byte(key)) != 1 {
			Error(w, &APIError{401, "internal authentication required"})
			return
		}
		if _, err := uuid.Parse(Owner(r)); err != nil {
			Error(w, &APIError{401, "user identity required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
func Observe(role string, next http.Handler) http.Handler {
	var requests atomic.Uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := uuid.NewString()
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		defer func() {
			if v := recover(); v != nil {
				slog.Error("panic", "error", v, "stack", string(debug.Stack()))
				Error(w, fmt.Errorf("panic: %v", v))
			}
			slog.Info("http request", "service", role, "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds(), "request_id", id)
		}()
		if r.URL.Path == "/metrics" {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			fmt.Fprintf(w, "# TYPE cargos_http_requests_total counter\ncargos_http_requests_total{service=%q} %d\n", role, requests.Load())
			return
		}
		requests.Add(1)
		next.ServeHTTP(w, r)
	})
}

// List and Get adapt typed read use cases to the common REST response envelope.
func List[T any](read func(context.Context, string, int, int) ([]T, error)) http.HandlerFunc {
	return Handle(func(w http.ResponseWriter, r *http.Request) error {
		limit, offset, err := Page(r)
		if err != nil {
			return err
		}
		items, err := read(r.Context(), Owner(r), limit, offset)
		if err != nil {
			return err
		}
		JSON(w, 200, map[string]any{"items": items, "limit": limit, "offset": offset})
		return nil
	})
}
func Get[T any](read func(context.Context, string, string) (*T, error)) http.HandlerFunc {
	return Handle(func(w http.ResponseWriter, r *http.Request) error {
		v, err := read(r.Context(), Owner(r), r.PathValue("id"))
		if err != nil {
			return err
		}
		JSON(w, 200, v)
		return nil
	})
}
