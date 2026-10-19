package renewal

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	payments "github.com/smartseaman/mobile-api/internal/domain/payment"
)

type handlerPaymentStore struct{ payments.Store }

func (s *handlerPaymentStore) Request(_ context.Context, id, user string) (payments.Request, error) {
	if user != "owner" {
		return payments.Request{}, payments.ErrNotFound
	}
	return payments.Request{ID: id, UserID: user}, nil
}
func (s *handlerPaymentStore) Current(context.Context, string) (payments.Attempt, error) {
	return payments.Attempt{ID: "transaction", RequestID: "request", RequestNo: "REQ001", UserID: "owner", Status: "SUCCESS", Input: payments.Input{PaymentMethod: "promptpay"}, Amount: 18000}, nil
}
func (s *handlerPaymentStore) ByKey(context.Context, string, string) (payments.Attempt, error) {
	return payments.Attempt{ID: "transaction", RequestID: "request", UserID: "owner", Status: "SUCCESS", Input: payments.Input{PaymentMethod: "promptpay"}}, nil
}

type handlerPaymentProvider struct{ payments.Provider }

func (p *handlerPaymentProvider) Ready() bool                                 { return true }
func (p *handlerPaymentProvider) VerifySignature(string, string, string) bool { return false }
func TestPaymentHandlerContract(t *testing.T) {
	service := NewService(nil, payments.NewService(&handlerPaymentStore{}, &handlerPaymentProvider{}, false))
	handler := NewHandler(service)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user.mobile_uuid", c.Get("Test-User", "owner")); return c.Next() })
	app.Get("/v1/document-renewals/:requestId/payments", handler.CurrentPayment)
	app.Post("/v1/document-renewals/:requestId/payments", handler.Payment)
	app.Post("/webhook", handler.Webhook)
	cases := []struct {
		method, path, body, user string
		status                   int
		code                     string
	}{
		{"GET", "/v1/document-renewals/request/payments", "", "owner", 200, "MA00000"},
		{"GET", "/v1/document-renewals/request/payments", "", "attacker", 404, "PAYMENT_NOT_FOUND"},
		{"POST", "/v1/document-renewals/request/payments", `{"paymentMethod":"promptpay"}`, "owner", 400, "PAYMENT_INVALID_INPUT"},
		{"POST", "/v1/document-renewals/request/payments", `{"paymentMethod":"promptpay","idempotencyKey":"old","returnUri":"app://changed"}`, "owner", 409, "PAYMENT_IDEMPOTENCY_CONFLICT"},
	}
	for _, test := range cases {
		req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Test-User", test.user)
		response, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		var payload struct {
			Code string            `json:"code"`
			Data payments.Response `json:"data"`
		}
		err = json.NewDecoder(response.Body).Decode(&payload)
		response.Body.Close()
		if err != nil || response.StatusCode != test.status || payload.Code != test.code {
			t.Fatalf("%s: status=%d,payload=%+v,error=%v", test.path, response.StatusCode, payload, err)
		}
		if test.status == 200 && (payload.Data.NextAction != "COMPLETE" || payload.Data.ServerTime == "") {
			t.Fatalf("missing checkout fields: %+v", payload)
		}
	}
	response, err := app.Test(httptest.NewRequest("POST", "/webhook", strings.NewReader(`{"key":"charge.complete","data":{"id":"charge"}}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 401 {
		t.Fatalf("webhook status=%d", response.StatusCode)
	}
}
