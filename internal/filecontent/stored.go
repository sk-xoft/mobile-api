// Package filecontent reads supported binary files and legacy Base64 objects.
package filecontent

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

// DecodeStored detects the payload rather than trusting legacy S3 content types.
// New uploads remain binary; this does not rewrite stored objects.
func DecodeStored(content []byte) ([]byte, string, error) {
	if mime := mediaType(content); mime != "" {
		return content, mime, nil
	}
	encoded := strings.TrimSpace(string(content))
	if strings.HasPrefix(encoded, "data:") {
		before, after, ok := strings.Cut(encoded, ",")
		if !ok || !strings.HasSuffix(before, ";base64") {
			return nil, "", fmt.Errorf("invalid stored file encoding")
		}
		encoded = after
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", fmt.Errorf("invalid stored file encoding")
	}
	if mime := mediaType(decoded); mime != "" {
		return decoded, mime, nil
	}
	return nil, "", fmt.Errorf("unsupported stored file content")
}

func mediaType(content []byte) string {
	switch {
	case bytes.HasPrefix(content, []byte("\x89PNG\r\n\x1a\n")):
		return "image/png"
	case bytes.HasPrefix(content, []byte("\xff\xd8\xff")):
		return "image/jpeg"
	case bytes.HasPrefix(content, []byte("%PDF-")):
		return "application/pdf"
	default:
		return ""
	}
}
