package middleware

import (
	"strings"
	"testing"
)

func TestSanitizeRequestBodyMasksBase64File(t *testing.T) {
	body := `{"fileCert":"data:application/pdf;base64,very-large-file","name":"certificate"}`
	result := sanitizeRequestBody("application/json", []byte(body))
	if result != `{"fileCert":"[FILE_BASE64_OMITTED]","name":"certificate"}` {
		t.Fatalf("sanitized body = %s", result)
	}
}

func TestSanitizeRequestBodyRedactsSecret(t *testing.T) {
	result := sanitizeRequestBody("application/json", []byte(`{"password":"do-not-log"}`))
	if result != `{"password":"[REDACTED]"}` {
		t.Fatalf("sanitized body = %s", result)
	}
}

func TestSanitizeRequestBodyOmitsMultipart(t *testing.T) {
	result := sanitizeRequestBody("multipart/form-data; boundary=test", []byte("binary content"))
	if result != "[MULTIPART_FILE_OMITTED]" {
		t.Fatalf("sanitized body = %s", result)
	}
}

func TestSanitizeResponseBodyOmitsBinary(t *testing.T) {
	result := sanitizeResponseBody("application/pdf", []byte("pdf content"))
	if result != "[BINARY_RESPONSE_OMITTED]" {
		t.Fatalf("sanitized body = %s", result)
	}
}

func TestCapLogBody(t *testing.T) {
	result := capLogBody(make([]byte, requestLogMaxBodyBytes+10))
	if len(result) != requestLogMaxBodyBytes+len("...[BODY_TRUNCATED]") {
		t.Fatalf("capped body length = %d", len(result))
	}
}

func TestSanitizeAuthCredentials(t *testing.T) {
	for _, body := range []string{`{"refToken":"refresh-secret"}`, `{"uid":"email-action-secret"}`} {
		result := sanitizeResponseBody("application/json", []byte(body))
		if strings.Contains(result, "secret") || !strings.Contains(result, redactedMarker) {
			t.Fatalf("credential leaked: %s", result)
		}
	}
}
