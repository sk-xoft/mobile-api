package payment_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	domain "github.com/smartseaman/mobile-api/internal/domain/payment"
	"github.com/smartseaman/mobile-api/internal/domain/renewal"
	provider "github.com/smartseaman/mobile-api/internal/platform/payment"
)

func TestPaymentHTTPProviderAuthErrorReturns503AndKeepsAttempt(t *testing.T) {
	for _, code := range []int{401, 403} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			fake := newFakeOmise()
			fake.postStatus = code
			service, store := setup(t, fake)
			handler := renewal.NewHandler(renewal.NewService(nil, service))
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error { c.Locals("user.mobile_uuid", "owner"); return c.Next() })
			app.Post("/payments/:requestId", handler.Payment)
			app.Get("/payments/:requestId", handler.CurrentPayment)
			app.Get("/payments/:requestId/:transactionId", handler.PaymentStatus)
			check := func(method, path, body string) {
				t.Helper()
				request := httptest.NewRequest(method, path, strings.NewReader(body))
				request.Header.Set("Content-Type", "application/json")
				response, err := app.Test(request)
				if err != nil {
					t.Fatal(err)
				}
				defer response.Body.Close()
				var payload map[string]any
				if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				if response.StatusCode != 503 || payload["code"] != "PAYMENT_UNAVAILABLE" || response.Header.Get("Retry-After") != "30" {
					t.Fatalf("%s %s: status=%d body=%+v retry=%s", method, path, response.StatusCode, payload, response.Header.Get("Retry-After"))
				}
				if _, exists := payload["data"]; exists {
					t.Fatal("provider auth failure returned a misleading payment snapshot")
				}
			}
			check("POST", "/payments/request", `{"paymentMethod":"promptpay","idempotencyKey":"one"}`)
			attempt, err := store.Current(context.Background(), "request")
			if err != nil {
				t.Fatal(err)
			}
			// These requests hit the same process during its provider-auth backoff.
			check("GET", "/payments/request", "")
			check("GET", "/payments/request/"+attempt.ID, "")
			check("POST", "/payments/request", `{"paymentMethod":"promptpay","idempotencyKey":"one"}`)
			check("POST", "/payments/request", `{"paymentMethod":"promptpay","idempotencyKey":"replacement"}`)
			store.mu.Lock()
			defer store.mu.Unlock()
			if len(store.attempts) != 1 || store.active != attempt.ID || store.paid != "" || store.transitions != 0 {
				t.Fatalf("auth failure changed payment state: attempts=%d active=%s paid=%s transitions=%d", len(store.attempts), store.active, store.paid, store.transitions)
			}
			stored := store.attempts[attempt.ID]
			if stored.Creation != "SENT" || stored.Charge.ID != "" || stored.Status != "PROCESSING" || stored.Review || stored.Lease != "" || fake.count() != 1 {
				t.Fatalf("attempt must remain uncertain with lease released: %+v, POSTs=%d", stored, fake.count())
			}
		})
	}
}

func TestExistingSentAttemptSurvivesAuthFailureAndRecoversWithoutNewCharge(t *testing.T) {
	fake := newFakeOmise()
	fake.loseResponse = true
	var reject atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reject.Load() {
			w.WriteHeader(401)
			return
		}
		fake.serve(w, r)
	}))
	defer server.Close()
	store := newMemoryStore()
	newService := func() *domain.Service {
		return domain.NewService(store, provider.NewOmise(server.URL, "test-key", "", 15*time.Minute), false)
	}
	service := newService()
	ctx := context.Background()
	first, err := service.Create(ctx, "owner", "request", input("one"))
	if err != nil || first.Status != "PROCESSING" {
		t.Fatalf("lost-response setup: payment=%+v error=%v", first, err)
	}
	reject.Store(true)
	_, err = service.Current(ctx, "owner", "request")
	assertCode(t, err, "PAYMENT_UNAVAILABLE")
	_, err = service.Get(ctx, "owner", "request", first.TransactionID)
	assertCode(t, err, "PAYMENT_UNAVAILABLE")
	if fake.count() != 1 {
		t.Fatal("auth failure caused another charge creation")
	}
	// Restarting after config repair clears only the in-memory provider backoff.
	reject.Store(false)
	recovered, err := newService().Current(ctx, "owner", "request")
	if err != nil || recovered.TransactionID != first.TransactionID || recovered.ProviderChargeID == "" || recovered.NextAction != "WAIT_FOR_PAYMENT" || fake.count() != 1 {
		t.Fatalf("recovery must find the original charge: payment=%+v error=%v posts=%d", recovered, err, fake.count())
	}
}

type unavailableAfterBackoff struct{ domain.Provider }

func (*unavailableAfterBackoff) Ready() bool                     { return true }
func (*unavailableAfterBackoff) AuthenticationUnavailable() bool { return true }

func TestKnownAuthFailureDoesNotBecomeSuccessWhenLeaseIsBusy(t *testing.T) {
	store := newMemoryStore()
	store.active = "attempt"
	store.attempts["attempt"] = domain.Attempt{ID: "attempt", RequestID: "request", UserID: "owner", Status: "PROCESSING", Creation: "SENT", Lease: "another-worker"}
	service := domain.NewService(store, &unavailableAfterBackoff{}, false)
	_, err := service.Current(context.Background(), "owner", "request")
	assertCode(t, err, "PAYMENT_UNAVAILABLE")
	if store.attempts["attempt"].Lease != "another-worker" {
		t.Fatal("read stole the active reconciliation lease")
	}
}
