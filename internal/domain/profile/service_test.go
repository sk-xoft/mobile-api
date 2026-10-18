package profile

import (
	"encoding/base64"
	"testing"
)

func TestDecodeImageAcceptsPNGDataURL(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0}
	value := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)

	content, contentType, err := decodeImage(value)
	if err != nil {
		t.Fatalf("decodeImage() error = %v", err)
	}
	if contentType != "image/png" || len(content) != len(png) {
		t.Fatalf("unexpected image: type=%q length=%d", contentType, len(content))
	}
}

func TestDecodeImageRejectsMismatchedDeclaredType(t *testing.T) {
	jpeg := []byte{0xff, 0xd8, 0xff, 0}
	value := "data:image/png;base64," + base64.StdEncoding.EncodeToString(jpeg)

	if _, _, err := decodeImage(value); err == nil {
		t.Fatal("decodeImage() expected a MIME mismatch error")
	}
}

func TestDecodeImageRejectsUnsupportedContent(t *testing.T) {
	value := base64.StdEncoding.EncodeToString([]byte("not an image"))
	if _, _, err := decodeImage(value); err == nil {
		t.Fatal("decodeImage() expected an unsupported content error")
	}
}
