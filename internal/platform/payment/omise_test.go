package paymentprovider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	domain "github.com/smartseaman/mobile-api/internal/domain/payment"
)

func TestHTTPTimeoutAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	client := NewOmise(server.URL, "fake-key", "", time.Minute)
	client.Client.Timeout = 20 * time.Millisecond
	start := time.Now()
	_, err := client.Retrieve(context.Background(), "charge")
	if err == nil || time.Since(start) > time.Second {
		t.Fatalf("timeout error=%v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Retrieve(ctx, "charge")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error=%v", err)
	}
}
func TestDecodeChargeAcceptsExpandedAndStringTransaction(t *testing.T) {
	for _, raw := range []string{`{"id":"charge","transaction":"transaction"}`, `{"id":"charge","transaction":{"id":"transaction"}}`} {
		c, err := decodeCharge([]byte(raw))
		if err != nil || c.TransactionID != "transaction" {
			t.Fatalf("charge=%+v,error=%v", c, err)
		}
	}
}
func TestWebhookWithConfiguredSecretRejectsMissingOrInvalidSignature(t *testing.T) {
	client := NewOmise("https://example.invalid", "fake-key", "secret", time.Minute)
	if client.VerifySignature("{}", "", "") || client.VerifySignature("{}", "invalid", "123") {
		t.Fatal("signature accepted")
	}
}

func TestProviderAuthFailureIsTypedRedactedAndBackedOff(t *testing.T) {
	for _, code := range []int{401, 403} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			var calls atomic.Int32
			var status atomic.Int32
			status.Store(int32(code))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(int(status.Load()))
				if status.Load() == 200 {
					_, _ = w.Write([]byte(`{"id":"charge","currency":"THB"}`))
				} else {
					_, _ = w.Write([]byte(`{"detail":"private-provider-response"}`))
				}
			}))
			defer server.Close()
			client := NewOmise(server.URL, "private-configured-key", "", time.Minute)
			_, err := client.Retrieve(context.Background(), "charge")
			var providerError *domain.ProviderHTTPError
			if !errors.As(err, &providerError) || providerError.StatusCode != code {
				t.Fatalf("expected typed HTTP %d error, got %v", code, err)
			}
			if strings.Contains(err.Error(), "private-") || client.Ready() || !client.AuthenticationUnavailable() {
				t.Fatal("provider failure leaked private data or did not pause calls")
			}
			var wg sync.WaitGroup
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := client.Retrieve(context.Background(), "charge")
					var unavailable *domain.Error
					if !errors.As(err, &unavailable) || unavailable.Code != "PAYMENT_UNAVAILABLE" {
						t.Errorf("expected unavailability during backoff, got %v", err)
					}
				}()
			}
			wg.Wait()
			if calls.Load() != 1 {
				t.Fatalf("requests during backoff=%d, want 1 total", calls.Load())
			}
			status.Store(200)
			client.authFailureUntil.Store(time.Now().Add(-time.Second).UnixNano())
			if !client.Ready() {
				t.Fatal("provider stayed unavailable after backoff elapsed")
			}
			if !client.AuthenticationUnavailable() {
				t.Fatal("known auth failure cleared before a successful probe")
			}
			if _, err := client.Retrieve(context.Background(), "charge"); err != nil || calls.Load() != 2 {
				t.Fatalf("probe after backoff: error=%v calls=%d", err, calls.Load())
			}
			if client.AuthenticationUnavailable() {
				t.Fatal("successful provider response did not clear auth failure")
			}
		})
	}
}
