package renewal

import (
	"database/sql"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestRenewalStatusAndProgressMapping(t *testing.T) {
	tests := map[string]string{
		"successful": "SUCCESS",
		"failed":     "FAILED",
		"expired":    "EXPIRED",
		"pending":    "PENDING",
		"unknown":    "PROCESSING",
	}
	for provider, expected := range tests {
		if got := omisePaymentStatus(provider); got != expected {
			t.Errorf("omisePaymentStatus(%q) = %q, want %q", provider, got, expected)
		}
	}
	for code, expected := range map[string]int{
		"DOCUMENT_REVIEW":            1,
		"MARINE_DEPARTMENT_RESULT":   2,
		"DEPARTMENT_DOCUMENT_PICKUP": 3,
		"DELIVERING":                 4,
		"DELIVERED":                  5,
		"UNKNOWN":                    0,
	} {
		if got := progress(code); got != expected {
			t.Errorf("progress(%q) = %d, want %d", code, got, expected)
		}
	}
}

func TestValidateItemStatus(t *testing.T) {
	tests := []struct {
		name     string
		item     ValidateItem
		files    []ValidateFile
		expected string
	}{
		{"missing", ValidateItem{}, nil, "MISSING"},
		{"needs fix", ValidateItem{}, []ValidateFile{{CheckResult: "FIX"}}, "NEED_FIX"},
		{"id card front and back", ValidateItem{ItemCode: "MRI001"}, []ValidateFile{{DocumentType: "ID_CARD", SlotCode: "FRONT"}, {DocumentType: "ID_CARD", SlotCode: "BACK"}}, "COMPLETE"},
		{"id card main", ValidateItem{ItemCode: "MRI001"}, []ValidateFile{{DocumentType: "ID_CARD", SlotCode: "MAIN"}}, "COMPLETE"},
		{"passport main", ValidateItem{ItemCode: "MRI001"}, []ValidateFile{{DocumentType: "PASSPORT", SlotCode: "MAIN"}}, "COMPLETE"},
		{"id card incomplete", ValidateItem{ItemCode: "MRI001"}, []ValidateFile{{DocumentType: "ID_CARD", SlotCode: "FRONT"}}, "INCOMPLETE"},
		{"general main", ValidateItem{}, []ValidateFile{{DocumentType: "GENERAL", SlotCode: "MAIN"}}, "COMPLETE"},
		{"general incomplete", ValidateItem{}, []ValidateFile{{DocumentType: "GENERAL", SlotCode: "FRONT"}}, "INCOMPLETE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validateItemStatus(test.item, test.files); got != test.expected {
				t.Errorf("validateItemStatus() = %q, want %q", got, test.expected)
			}
		})
	}
}

func TestRenewalHelpersFormatNullableValues(t *testing.T) {
	if stringPointer("") != nil || stringPointer("value") == nil || *stringPointer("value") != "value" {
		t.Fatal("stringPointer() did not preserve optional values")
	}
	if nullString("") != nil || nullString("value") != "value" {
		t.Fatal("nullString() did not normalize values")
	}
	when := time.Date(2026, time.October, 18, 8, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	if nullableTime(time.Time{}) != nil {
		t.Fatal("zero time should become nil")
	}
	if got := formatTime(when); got != "2026-10-18T01:00:00Z" {
		t.Fatalf("formatTime() = %q", got)
	}
	if got := formatNullableTime(sql.NullTime{Time: when, Valid: true}); got != "2026-10-18T01:00:00Z" {
		t.Fatalf("formatNullableTime() = %q", got)
	}
}

func TestOmiseCreateChargeSendsConfiguredPromptPayExpiry(t *testing.T) {
	var received url.Values
	client := newOmiseClient(15 * time.Minute)
	client.secretKey = "secret"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/charges" || r.Header.Get("Idempotency-Key") != "idem-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		body, _ := io.ReadAll(r.Body)
		received, _ = url.ParseQuery(string(body))
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"chrg_test","status":"pending","expires_at":"2026-10-18T01:15:00Z","source":{"id":"src_test","scannable_code":{"image":{"download_uri":"https://qr.test"}}}}`))}, nil
	})}
	client.baseURL = "https://omise.test"
	charge, err := client.CreateCharge(10, "REQ001", PaymentRequest{PaymentMethod: "promptpay", IdempotencyKey: "idem-1"})
	if err != nil {
		t.Fatal(err)
	}
	if charge.ExpiresAt.IsZero() || charge.QRCodeURI != "https://qr.test" {
		t.Fatalf("unexpected charge: %#v", charge)
	}
	if received.Get("expires_at") == "" || !strings.HasPrefix(received.Get("expires_at"), "20") {
		t.Fatalf("expires_at was not sent: %q", received.Get("expires_at"))
	}
}

func TestOmiseSignatureVerification(t *testing.T) {
	client := newOmiseClient()
	client.webhookSecret = "secret"
	if client.VerifySignature("payload", "", "") {
		t.Fatal("missing signature should fail")
	}
	client.webhookSecret = ""
	if !client.VerifySignature("payload", "", "") {
		t.Fatal("disabled signature verification should pass")
	}
}
