package document

import (
	"encoding/base64"
	"errors"
	"testing"
)

func TestDecodeDocumentAcceptsSupportedDataURLs(t *testing.T) {
	tests := []struct {
		name, contentType string
		content           []byte
	}{
		{"png", "image/png", []byte("\x89PNG\r\n\x1a\n")},
		{"jpeg", "image/jpeg", []byte("\xff\xd8\xff\xe0")},
		{"pdf", "application/pdf", []byte("%PDF-1.7")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := "data:" + test.contentType + ";base64," + base64.StdEncoding.EncodeToString(test.content)
			decoded, contentType, err := decodeDocument(value)
			if err != nil || contentType != test.contentType || string(decoded) != string(test.content) {
				t.Fatalf("decodeDocument() = (%q, %q, %v)", decoded, contentType, err)
			}
		})
	}
}

func TestDecodeDocumentRejectsMismatchedContent(t *testing.T) {
	value := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("not-an-image"))
	_, _, err := decodeDocument(value)
	var businessError *BusinessError
	if !errors.As(err, &businessError) || businessError.Code != "MA00005" {
		t.Fatalf("expected MA00005, got %v", err)
	}
}

func TestValidateDates(t *testing.T) {
	if err := validateDates("2026-01-01", "9999-99-99"); err != nil {
		t.Fatalf("open-ended date should be accepted: %v", err)
	}
	if err := validateDates("2026-02-30", "2027-01-01"); err == nil {
		t.Fatal("invalid start date should be rejected")
	}
}
