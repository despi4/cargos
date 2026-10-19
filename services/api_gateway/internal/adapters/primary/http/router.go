package http

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/depsi4/cargos/api_gateway/docs"
	"github.com/depsi4/cargos/internal/platform/httpx"
	"github.com/depsi4/cargos/pkg/fault"
	"github.com/google/uuid"
)

func NewRouter(mux *http.ServeMux, urls map[string]string, key string) http.Handler {
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}).DialContext, ResponseHeaderTimeout: 50 * time.Second, MaxIdleConns: 100, MaxIdleConnsPerHost: 20, IdleConnTimeout: 60 * time.Second}
	client := &http.Client{Timeout: 5 * time.Second, Transport: transport}
	mux.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.FS(docs.Files))))
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(docs.Spec)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, 200, map[string]string{"name": "Cargos API", "version": "1.0.0", "docs": "/docs/", "openapi": "/openapi.json", "health": "/readyz"})
	})
	mux.HandleFunc("GET /readyz", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		checks := make(chan bool, len(urls))
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		for _, addr := range urls {
			go func(addr string) {
				req, err := http.NewRequestWithContext(ctx, "GET", addr+"/readyz", nil)
				if err != nil {
					checks <- false
					return
				}
				resp, err := client.Do(req)
				if err != nil {
					checks <- false
					return
				}
				resp.Body.Close()
				checks <- resp.StatusCode == 200
			}(addr)
		}
		healthy := true
		for range urls {
			if !<-checks {
				healthy = false
			}
		}
		if !healthy {
			return &httpx.APIError{Status: 503, Message: "one or more services are unavailable"}
		}
		httpx.JSON(w, 200, map[string]string{"status": "ready"})
		return nil
	}))
	routes := map[string]string{"auth": "auth", "products": "inventory", "movements": "inventory", "suppliers": "procurement", "orders": "procurement", "demand": "demand", "predictions": "prediction", "notifications": "notification"}
	proxies := map[string]*httputil.ReverseProxy{}
	for role, addr := range urls {
		target, _ := url.Parse(addr)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = transport
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			httpx.Error(w, &httpx.APIError{Status: 503, Message: "upstream service unavailable"})
		}
		proxies[role] = proxy
	}
	limiter := newAuthLimiter()
	mux.Handle("/api/v1/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/v1/")
		segment := strings.Split(rest, "/")[0]
		role, ok := routes[segment]
		if !ok {
			httpx.Error(w, fault.ErrNotFound)
			return
		}
		// Never trust identity or forwarding headers supplied by the caller.
		r.Header.Del("X-Internal-Key")
		r.Header.Del("X-User-ID")
		r.Header.Del("Forwarded")
		r.Header.Del("X-Forwarded-For")
		r.Header.Del("X-Forwarded-Host")
		r.Header.Del("X-Forwarded-Proto")
		if role == "auth" {
			if r.Method == "POST" && r.URL.Path != "/api/v1/auth/logout" && !limiter.allow(r.RemoteAddr) {
				w.Header().Set("Retry-After", "60")
				httpx.Error(w, &httpx.APIError{Status: 429, Message: "too many authentication attempts; retry in one minute"})
				return
			}
		} else {
			req, err := http.NewRequestWithContext(r.Context(), "GET", urls["auth"]+"/api/v1/auth/me", nil)
			if err != nil {
				httpx.Error(w, err)
				return
			}
			req.Header.Set("Authorization", r.Header.Get("Authorization"))
			resp, err := client.Do(req)
			if err != nil {
				httpx.Error(w, &httpx.APIError{Status: 503, Message: "authentication service unavailable"})
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				if resp.StatusCode == 401 {
					httpx.Error(w, &httpx.APIError{Status: 401, Message: "valid Bearer token required"})
				} else {
					httpx.Error(w, &httpx.APIError{Status: 503, Message: "authentication service unavailable"})
				}
				return
			}
			var user struct {
				ID string `json:"id"`
			}
			if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&user); err != nil {
				httpx.Error(w, err)
				return
			}
			if _, err = uuid.Parse(user.ID); err != nil {
				httpx.Error(w, &httpx.APIError{Status: 503, Message: "invalid upstream identity"})
				return
			}
			r.Header.Set("X-Internal-Key", key)
			r.Header.Set("X-User-ID", user.ID)
			r.Header.Del("Authorization")
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		proxies[role].ServeHTTP(w, r)
	}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { httpx.Error(w, fault.ErrNotFound) })
	return mux
}

type authWindow struct {
	start time.Time
	count int
}
type authLimiter struct {
	mu      sync.Mutex
	entries map[string]authWindow
}

func newAuthLimiter() *authLimiter { return &authLimiter{entries: map[string]authWindow{}} }
func (l *authLimiter) allow(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range l.entries {
		if now.Sub(v.start) > time.Minute {
			delete(l.entries, k)
		}
	}
	v := l.entries[host]
	if v.start.IsZero() {
		v.start = now
	}
	v.count++
	l.entries[host] = v
	return v.count <= 30
}
