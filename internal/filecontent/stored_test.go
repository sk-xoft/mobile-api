package filecontent

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestDecodeStoredBinaryAndLegacy(t *testing.T) {
	for _, file := range []struct {
		mime string
		body []byte
	}{
		{"image/png", []byte("\x89PNG\r\n\x1a\n\x00")},
		{"image/jpeg", []byte("\xff\xd8\xff\x00")},
		{"application/pdf", []byte("%PDF-1.7\n")},
	} {
		for _, encoding := range []string{"binary", "base64", "data-url"} {
			t.Run(file.mime+"/"+encoding, func(t *testing.T) {
				stored := file.body
				if encoding == "base64" {
					stored = []byte("\n" + base64.StdEncoding.EncodeToString(file.body) + "\n")
				}
				if encoding == "data-url" {
					stored = []byte("data:" + file.mime + ";base64," + base64.StdEncoding.EncodeToString(file.body))
				}
				got, mime, err := DecodeStored(stored)
				if err != nil || mime != file.mime || !bytes.Equal(got, file.body) {
					t.Fatalf("mime=%q length=%d err=%v", mime, len(got), err)
				}
			})
		}
	}
}
func TestDecodeStoredRejectsInvalidPayloads(t *testing.T) {
	for _, stored := range []string{"", "not-base64!", "aGVsbG8=", "data:image/png,raw", "data:image/png;base64,%%%"} {
		if _, _, err := DecodeStored([]byte(stored)); err == nil {
			t.Fatal("expected invalid stored content error")
		}
	}
}
