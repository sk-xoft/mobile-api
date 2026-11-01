package document

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestViewCertificateServesLegacyAndBinary(t *testing.T) {
	pdf := []byte("%PDF-1.7\n")
	for _, legacy := range []bool{true, false} {
		t.Run(map[bool]string{true: "legacy", false: "binary"}[legacy], func(t *testing.T) {
			body := pdf
			if legacy {
				body = []byte(base64.StdEncoding.EncodeToString(pdf))
			}
			repo := &fakeDocumentRepository{certificate: Certificate{FileName: "stored"}}
			storage := &fakeDocumentStorage{getContent: body, getType: "text/plain"}
			handler := NewHandler(NewService(repo, storage, "documents/%s/certs/%s"))
			app := fiber.New()
			app.Get("/view", func(c *fiber.Ctx) error { c.Locals("user.mobile_uuid", "user-1"); return handler.View(c) })
			res, err := app.Test(httptest.NewRequest("GET", "/view?certCode=DOC001", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			got, _ := io.ReadAll(res.Body)
			if res.StatusCode != 200 || res.Header.Get("Content-Type") != "application/pdf" || !bytes.Equal(got, pdf) {
				t.Fatalf("status=%d type=%s length=%d", res.StatusCode, res.Header.Get("Content-Type"), len(got))
			}
		})
	}
}
