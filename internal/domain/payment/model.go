package payment

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("payment not found")

const ProviderAuthRetryAfterSeconds = 30

// ProviderHTTPError carries status only: provider bodies may contain sensitive data.
type ProviderHTTPError struct {
	StatusCode int
}

func (e *ProviderHTTPError) Error() string {
	return fmt.Sprintf("payment provider HTTP %d", e.StatusCode)
}

type Error struct {
	Code    string
	Payment *Response
}

func (e *Error) Error() string   { return e.Code }
func Conflict(code string) error { return &Error{Code: code} }

type Input struct {
	PaymentMethod  string `json:"paymentMethod"`
	IdempotencyKey string `json:"idempotencyKey"`
	ReturnURI      string `json:"returnUri"`
	PlatformType   string `json:"platformType"`
}
type Request struct {
	ID, No, UserID, Status string
	Amount                 int64
	Active                 bool
}
type Charge struct {
	ID                                                                                            string    `json:"id"`
	Reference                                                                                     string    `json:"ref_id"`
	RequestNo                                                                                     string    `json:"request_no"`
	Status                                                                                        string    `json:"status"`
	Amount                                                                                        int64     `json:"amount"`
	Currency                                                                                      string    `json:"currency"`
	Live                                                                                          bool      `json:"livemode"`
	Expired                                                                                       bool      `json:"expired"`
	ExpiresAt                                                                                     time.Time `json:"expires_at"`
	SourceID, TransactionID, AuthorizeURI, ReturnURI, QRCodeURI, FailureCode, FailureMessage, Raw string
}
type Attempt struct {
	ID, No, RequestID, RequestNo, UserID, Key, Fingerprint, Reference, Status, Creation string
	Input                                                                               Input
	Amount                                                                              int64
	Charge                                                                              Charge
	CreatedAt                                                                           time.Time
	SentAt                                                                              time.Time
	CheckedAt                                                                           time.Time
	Review                                                                              bool
	Paid                                                                                bool
	Legacy                                                                              bool
	Failures                                                                            int
	Lease                                                                               string
}
type Response struct {
	RequestID         string  `json:"requestId"`
	RequestNo         string  `json:"requestNo"`
	TransactionID     string  `json:"transactionId"`
	TransactionNo     string  `json:"transactionNo"`
	Status            string  `json:"status"`
	Channel           string  `json:"channel"`
	PaymentMethod     string  `json:"paymentMethod"`
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	Provider          string  `json:"provider"`
	ProviderChargeID  string  `json:"providerChargeId,omitempty"`
	AuthorizeURI      string  `json:"authorizeUri,omitempty"`
	ReturnURI         string  `json:"returnUri,omitempty"`
	QRCodeURI         string  `json:"qrCodeDownloadUri,omitempty"`
	ExpiresAt         string  `json:"expiresAt,omitempty"`
	FailureCode       string  `json:"failureCode,omitempty"`
	FailureMessage    string  `json:"failureMessage,omitempty"`
	ServerTime        string  `json:"serverTime"`
	QRExpired         bool    `json:"qrExpired"`
	NextAction        string  `json:"nextAction"`
	RetryAfterSeconds int     `json:"retryAfterSeconds"`
}

type Store interface {
	Request(context.Context, string, string) (Request, error)
	ByKey(context.Context, string, string) (Attempt, error)
	Current(context.Context, string) (Attempt, error)
	Get(context.Context, string) (Attempt, error)
	Reserve(context.Context, Attempt) (Attempt, bool, error)
	Acquire(context.Context, string, string, bool) (bool, error)
	MarkSent(context.Context, Attempt) (bool, error)
	Apply(context.Context, Attempt, Charge) error
	Retry(context.Context, Attempt, bool) error
	Due(context.Context, int) ([]string, error)
	Find(context.Context, Charge) (Attempt, error)
}
type Provider interface {
	Ready() bool
	Create(context.Context, Attempt) (Charge, error)
	Retrieve(context.Context, string) (Charge, error)
	Find(context.Context, Attempt) (Charge, error)
	VerifySignature(string, string, string) bool
}

// ProviderAuthenticationState lets a cached response retain a known auth failure
// while another caller holds the reconciliation lease after backoff expires.
type ProviderAuthenticationState interface {
	AuthenticationUnavailable() bool
}
