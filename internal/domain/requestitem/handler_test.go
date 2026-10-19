package requestitem

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestItemHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, method, path string
		upload             bool
		repoErr            error
		want               int
		code               string
	}{
		{"list", "GET", "/items", false, nil, 200, "MA00000"},
		{"preview", "GET", "/items/A", false, nil, 200, "MA00000"},
		{"renewal preview", "GET", "/renewals/RN/A", false, nil, 200, "MA00000"},
		{"legacy renewal preview", "GET", "/renewal-items/A?request_no=RN", false, nil, 200, "MA00000"},
		{"upload", "POST", "/items/A", true, nil, 200, "MA00000"},
		{"renewal upload", "POST", "/renewals/RN/A", true, nil, 200, "MA00000"},
		{"missing file", "POST", "/items/A", false, nil, 400, "MA00005"},
		{"not found", "GET", "/renewals/RN/A", false, ErrNotFound, 400, "MA00016"},
		{"database error", "GET", "/items", false, errors.New("db"), 500, "MA00013"},
		{"business error", "GET", "/items", false, &BusinessError{Code: "MA00005", Detail: "documentStatus"}, 400, "MA00005"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRequestItemRepo{status: "PAYMENT_PENDING", err: tc.repoErr, renewalID: "r", itemID: "i", files: []File{{DocumentType: "GENERAL", SlotCode: "MAIN", UploadedAt: "now"}}}
			h := NewHandler(NewService(repo, &fakeRequestItemStorage{}, "uploads"))
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error { c.Locals("user.mobile_uuid", "user"); return c.Next() })
			app.Get("/items", h.List)
			app.Get("/items/:itemCode", h.Preview)
			app.Post("/items/:itemCode", h.Upload)
			app.Get("/renewals/:requestNo/:documentRequestItemCode", h.RenewalPreview)
			app.Get("/renewal-items/:documentRequestItemCode", h.RenewalPreview)
			app.Post("/renewals/:requestNo/:documentRequestItemCode", h.RenewalUpload)
			var body bytes.Buffer
			contentType := ""
			if tc.upload {
				writer := multipart.NewWriter(&body)
				part, err := writer.CreateFormFile("file", "f.pdf")
				if err != nil {
					t.Fatal(err)
				}
				if _, err = part.Write([]byte("%PDF-1.7")); err != nil {
					t.Fatal(err)
				}
				if err = writer.Close(); err != nil {
					t.Fatal(err)
				}
				contentType = writer.FormDataContentType()
			}
			req := httptest.NewRequest(tc.method, tc.path, &body)
			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			var envelope struct {
				Code string `json:"code"`
			}
			if err = json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.want || envelope.Code != tc.code {
				t.Fatalf("status=%d code=%s want %d %s", resp.StatusCode, envelope.Code, tc.want, tc.code)
			}
		})
	}
}
