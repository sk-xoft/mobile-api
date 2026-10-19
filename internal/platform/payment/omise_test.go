package paymentprovider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
