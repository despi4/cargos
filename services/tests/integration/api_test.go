package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	gatewayapp "github.com/depsi4/cargos/api_gateway"
	auth "github.com/depsi4/cargos/auth_service"
	demand "github.com/depsi4/cargos/demand_service"
	"github.com/depsi4/cargos/internal/platform/client"
	"github.com/depsi4/cargos/internal/platform/config"
	"github.com/depsi4/cargos/internal/platform/server"
	inventory "github.com/depsi4/cargos/inventory_service"
	notification "github.com/depsi4/cargos/notification_service"
	prediction "github.com/depsi4/cargos/prediction_service"
	procurement "github.com/depsi4/cargos/procurement_service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx := context.Background()
	key := strings.Repeat("k", 32)
	urls := map[string]string{}

	apps := map[string]server.Application{}
	servers := map[string]*httptest.Server{}
	cfg := config.Config{DatabaseURL: dsn, InternalKey: key, JWTSecret: strings.Repeat("s", 32), URLs: urls}
	factories := map[string]server.Factory{"auth": auth.New, "inventory": inventory.New, "procurement": procurement.New, "demand": demand.New, "prediction": prediction.New, "notification": notification.New}
	for _, role := range []string{"auth", "inventory", "procurement", "demand", "prediction", "notification"} {
		app, err := factories[role](ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(app.Close)
		apps[role] = app
		backend := httptest.NewServer(app.Handler)
		t.Cleanup(backend.Close)
		urls[role] = backend.URL
		servers[role] = backend
	}
	app, err := gatewayapp.New(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	gateway := httptest.NewServer(app.Handler)
	defer gateway.Close()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	transport := client.New(key, urls)
	token := ""
	send := func(method, path string, in any, status int, out any) {
		t.Helper()
		var body io.Reader
		if in != nil {
			data, _ := json.Marshal(in)
			body = bytes.NewReader(data)
		}
		req, _ := http.NewRequest(method, gateway.URL+path, body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-User-ID", uuid.NewString())
		req.Header.Set("X-Internal-Key", key)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != status {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, resp.StatusCode, status, data)
		}
		if out != nil {
			if err = json.Unmarshal(data, out); err != nil {
				t.Fatalf("decode %s: %v", data, err)
			}
		}
	}
	send("GET", "/readyz", nil, 200, nil)
	send("GET", "/openapi.json", nil, 200, nil)
	send("GET", "/docs/", nil, 200, nil)
	var account struct {
		ID string `json:"id"`
	}
	email := uuid.NewString() + "@example.com"
	credentials := map[string]any{"username": uuid.NewString(), "email": email, "password": "correct horse battery"}
	send("POST", "/api/v1/auth/register", credentials, 201, &account)
	send("POST", "/api/v1/auth/register", credentials, 409, nil)
	send("POST", "/api/v1/auth/login", map[string]string{"email": email, "password": "wrong"}, 401, nil)
	var tokens tokenPair
	send("POST", "/api/v1/auth/login", map[string]any{"email": email, "password": credentials["password"]}, 200, &tokens)
	token = tokens.Access
	var p Product
	send("POST", "/api/v1/products", ProductInput{"SKU-1", "Tyres", "pcs", 10, 50}, 201, &p)
	send("POST", "/api/v1/products", ProductInput{"SKU-1", "Tyres", "pcs", 10, 50}, 409, nil)
	var supplier Supplier
	send("POST", "/api/v1/suppliers", SupplierInput{"Supplier", "s@example.com", 7}, 201, &supplier)
	movePath := "/api/v1/products/" + p.ID + "/movements"
	movement := map[string]any{"quantity": 2, "reason": "opening", "reference": "opening"}
	send("POST", movePath, movement, 200, nil)
	send("POST", movePath, movement, 200, nil)
	movement["quantity"] = 3
	send("POST", movePath, movement, 409, nil)
	send("POST", movePath, map[string]any{"quantity": -3, "reason": "consume", "reference": "too-much"}, 409, nil)
	send("POST", "/api/v1/demand", DemandInput{p.ID, 90, time.Now().UTC(), "sales"}, 201, nil)
	var prediction Prediction
	send("POST", "/api/v1/predictions", map[string]any{"window_days": 30, "supplier_id": supplier.ID}, 201, &prediction)
	if len(prediction.Recommendations) != 1 || prediction.Recommendations[0].SuggestedQuantity != 48 {
		t.Fatalf("unexpected recommendations: %+v", prediction)
	}
	// A network outage leaves a durable pending event. A retry deduplicates delivery.
	notifyURL := urls["notification"]
	urls["notification"] = "http://127.0.0.1:1"
	apps["prediction"].Tick(ctx)
	urls["notification"] = notifyURL
	apps["prediction"].Tick(ctx)
	// Emulate a crash after receiver commit but before outbox acknowledgement.
	if _, err := admin.Exec(ctx, `UPDATE prediction.predictions SET notification_sent=false WHERE owner_id=$1 AND id=$2`, account.ID, prediction.ID); err != nil {
		t.Fatal(err)
	}
	apps["prediction"].Tick(ctx)
	var inbox struct {
		Items []Notification `json:"items"`
	}
	send("GET", "/api/v1/notifications", nil, 200, &inbox)
	if len(inbox.Items) != 1 {
		t.Fatalf("expected one notification, got %+v", inbox)
	}
	send("PATCH", "/api/v1/notifications/"+inbox.Items[0].ID, map[string]bool{"read": true}, 200, nil)
	var order Order
	send("POST", "/api/v1/orders", OrderInput{p.ID, supplier.ID, 48, 1000, "KZT"}, 201, &order)
	statusPath := "/api/v1/orders/" + order.ID + "/status"
	send("POST", statusPath, map[string]string{"status": "received"}, 409, nil)
	send("POST", statusPath, map[string]string{"status": "ordered"}, 200, nil)
	var incomingPrediction Prediction
	send("POST", "/api/v1/predictions", map[string]int{"window_days": 30}, 201, &incomingPrediction)
	if len(incomingPrediction.Recommendations) != 0 {
		t.Fatalf("incoming stock was ignored: %+v", incomingPrediction)
	}
	inventoryURL := urls["inventory"]
	urls["inventory"] = "http://127.0.0.1:1"
	send("POST", statusPath, map[string]string{"status": "received"}, 503, nil)
	send("POST", statusPath, map[string]string{"status": "cancelled"}, 409, nil)
	var interrupted Order
	send("GET", "/api/v1/orders/"+order.ID, nil, 200, &interrupted)
	if interrupted.Status != "receiving" {
		t.Fatalf("receipt intent not persisted: %+v", interrupted)
	}
	urls["inventory"] = inventoryURL
	// Emulate inventory committing just before the procurement process crashes.
	if err := transport.Do(ctx, account.ID, "inventory", "POST", "/internal/products/"+p.ID+"/receipts", map[string]any{"quantity": 48, "reason": "purchase received", "reference": "purchase:" + order.ID}, nil); err != nil {
		t.Fatal(err)
	}
	apps["procurement"].Tick(ctx)
	send("POST", statusPath, map[string]string{"status": "received"}, 200, nil)
	send("POST", statusPath, map[string]string{"status": "received"}, 200, nil)
	send("POST", statusPath, map[string]string{"status": "cancelled"}, 409, nil)
	send("GET", "/api/v1/products/"+p.ID, nil, 200, &p)
	if p.Stock != 50 {
		t.Fatalf("receipt must credit once: %d", p.Stock)
	}
	// Two racing withdrawals cannot oversell; the ledger and stock agree.
	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			data, _ := json.Marshal(map[string]any{"quantity": -40, "reason": "concurrent withdrawal", "reference": fmt.Sprintf("race-%d", i)})
			req, _ := http.NewRequest("POST", gateway.URL+movePath, bytes.NewReader(data))
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				statuses <- 0
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			statuses <- resp.StatusCode
		}(i)
	}
	wg.Wait()
	close(statuses)
	success, conflict := 0, 0
	for code := range statuses {
		switch code {
		case 200:
			success++
		case 409:
			conflict++
		default:
			t.Fatalf("unexpected concurrent status %d", code)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("overselling protection failed: %d success, %d conflicts", success, conflict)
	}
	send("GET", "/api/v1/products/"+p.ID, nil, 200, &p)
	if p.Stock != 10 {
		t.Fatalf("unexpected balance %d", p.Stock)
	}
	// Concurrent identical idempotency references must produce only one ledger entry.
	statuses = make(chan int, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", gateway.URL+movePath, strings.NewReader(`{"quantity":5,"reason":"retry","reference":"concurrent-retry"}`))
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				statuses <- 0
				return
			}
			resp.Body.Close()
			statuses <- resp.StatusCode
		}()
	}
	wg.Wait()
	close(statuses)
	for code := range statuses {
		if code != 200 {
			t.Fatalf("retry returned %d", code)
		}
	}
	send("GET", "/api/v1/products/"+p.ID, nil, 200, &p)
	if p.Stock != 15 {
		t.Fatalf("duplicate stock credit: %d", p.Stock)
	}
	// An independent account cannot read or mutate the first account's resources.
	firstToken := token
	other := uuid.NewString()
	send("POST", "/api/v1/auth/register", map[string]string{"username": other, "email": other + "@example.com", "password": "another-password"}, 201, nil)
	var second tokenPair
	send("POST", "/api/v1/auth/login", map[string]string{"email": other + "@example.com", "password": "another-password"}, 200, &second)
	token = second.Access
	send("GET", "/api/v1/products/"+p.ID, nil, 404, nil)
	send("POST", movePath, map[string]any{"quantity": 1, "reason": "spoof", "reference": "spoof"}, 404, nil)
	token = firstToken
	oldRefresh := tokens.Refresh
	send("POST", "/api/v1/auth/refresh", map[string]string{"refresh_token": oldRefresh}, 200, &tokens)
	send("GET", "/api/v1/auth/me", nil, 401, nil)
	send("POST", "/api/v1/auth/refresh", map[string]string{"refresh_token": oldRefresh}, 401, nil)
	token = tokens.Access
	var finalOrder Order
	send("POST", "/api/v1/orders", OrderInput{p.ID, supplier.ID, 1, 1000, "KZT"}, 201, &finalOrder)
	send("POST", "/api/v1/orders/"+finalOrder.ID+"/status", map[string]string{"status": "ordered"}, 200, nil)
	send("DELETE", "/api/v1/products/"+p.ID, nil, 204, nil)
	send("POST", "/api/v1/orders/"+finalOrder.ID+"/status", map[string]string{"status": "received"}, 200, nil)
	send("GET", "/api/v1/products/"+p.ID, nil, 200, &p)
	if p.Stock != 16 {
		t.Fatalf("existing purchase could not complete after archive: %d", p.Stock)
	}
	send("POST", "/internal/products/"+p.ID+"/receipts", map[string]any{"quantity": 1, "reason": "bypass", "reference": "bypass"}, 404, nil)
	send("POST", movePath, map[string]any{"quantity": 1, "reason": "archived", "reference": "archived"}, 409, nil)
	send("GET", "/api/v1/products?limit=0", nil, 400, nil)
	send("POST", "/api/v1/auth/logout", nil, 204, nil)
	send("GET", "/api/v1/products", nil, 401, nil)
	send("POST", "/api/v1/auth/refresh", map[string]string{"refresh_token": tokens.Refresh}, 401, nil)
	// A direct backend call cannot bypass gateway authentication.
	resp, err := http.Get(servers["inventory"].URL + "/api/v1/products")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Fatalf("direct backend access returned %d", resp.StatusCode)
	}
}
