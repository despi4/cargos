// Package server owns the process lifecycle, not service composition or business rules.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/depsi4/cargos/internal/platform/config"
	"github.com/depsi4/cargos/internal/platform/httpx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	Handler http.Handler
	Close   func()
	Tick    func(context.Context) error
}
type Factory func(context.Context, config.Config) (Application, error)

func Run(name string, newApp Factory) error {
	cfg, err := config.Load(name)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 30*time.Second)
	app, err := newApp(startup, cfg)
	cancel()
	if err != nil {
		return err
	}
	if app.Close != nil {
		defer app.Close()
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		if app.Tick == nil {
			return
		}
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := app.Tick(ctx); err != nil && ctx.Err() == nil {
					slog.Warn("background task will retry", "service", name, "error", err)
				}
			}
		}
	}()
	srv := &http.Server{Addr: cfg.Address, Handler: httpx.Observe(name, app.Handler), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	failure := make(chan error, 1)
	go func() {
		slog.Info("service started", "service", name, "addr", cfg.Address)
		failure <- srv.ListenAndServe()
	}()
	select {
	case err := <-failure:
		stop()
		<-done
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = srv.Shutdown(shutdown)
	<-done
	return err
}
func Health(mux *http.ServeMux, name string, pool *pgxpool.Pool) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, 200, map[string]string{"status": "ok", "service": name})
	})
	if pool == nil {
		return
	}
	mux.HandleFunc("GET /readyz", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return &httpx.APIError{Status: 503, Message: "database unavailable"}
		}
		httpx.JSON(w, 200, map[string]string{"status": "ready"})
		return nil
	}))
}
