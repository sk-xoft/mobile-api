package paymentprovider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	domain "github.com/smartseaman/mobile-api/internal/domain/payment"
)

type Omise struct {
	BaseURL, SecretKey, WebhookSecret string
	QRExpire                          time.Duration
	Client                            *http.Client
	authFailureUntil                  atomic.Int64
}

func NewOmise(base, key, webhook string, expiry time.Duration) *Omise {
	return &Omise{BaseURL: strings.TrimRight(base, "/"), SecretKey: key, WebhookSecret: webhook, QRExpire: expiry, Client: &http.Client{Timeout: 10 * time.Second}}
}
func (o *Omise) Ready() bool {
	return o.SecretKey != "" && time.Now().UnixNano() >= o.authFailureUntil.Load()
}
func (o *Omise) AuthenticationUnavailable() bool {
	return o.authFailureUntil.Load() != 0
}
func (o *Omise) Create(ctx context.Context, a domain.Attempt) (domain.Charge, error) {
	form := url.Values{"amount": {strconv.FormatInt(a.Amount, 10)}, "currency": {"THB"}, "description": {"Document renewal " + a.RequestNo}, "ref_id": {a.Reference}, "metadata[payment_attempt_id]": {a.Reference}, "metadata[request_no]": {a.RequestNo}, "source[type]": {a.Input.PaymentMethod}}
	if a.Input.ReturnURI != "" {
		form.Set("return_uri", a.Input.ReturnURI)
	}
	if a.Input.PlatformType != "" {
		form.Set("source[platform_type]", a.Input.PlatformType)
	}
	if a.Input.PaymentMethod == "promptpay" {
		form.Set("expires_at", time.Now().UTC().Add(o.QRExpire).Format(time.RFC3339))
	}
	raw, err := o.request(ctx, http.MethodPost, "/charges", form)
	if err != nil {
		return domain.Charge{}, err
	}
	return decodeCharge(raw)
}
func (o *Omise) Retrieve(ctx context.Context, id string) (domain.Charge, error) {
	raw, err := o.request(ctx, http.MethodGet, "/charges/"+url.PathEscape(id), nil)
	if err != nil {
		return domain.Charge{}, err
	}
	return decodeCharge(raw)
}
func (o *Omise) Find(ctx context.Context, a domain.Attempt) (domain.Charge, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	start := a.SentAt
	if start.IsZero() {
		start = a.CreatedAt
	}
	for offset := 0; offset < 2000; offset += 20 {
		values := url.Values{"from": {start.Add(-time.Minute).UTC().Format(time.RFC3339)}, "to": {start.Add(5 * time.Minute).UTC().Format(time.RFC3339)}, "limit": {"20"}, "offset": {strconv.Itoa(offset)}, "order": {"chronological"}}
		raw, err := o.request(ctx, http.MethodGet, "/charges?"+values.Encode(), nil)
		if err != nil {
			return domain.Charge{}, err
		}
		var list struct {
			Data []json.RawMessage `json:"data"`
		}
		if err = json.Unmarshal(raw, &list); err != nil {
			return domain.Charge{}, err
		}
		for _, rawCharge := range list.Data {
			c, err := decodeCharge(rawCharge)
			if err != nil {
				return domain.Charge{}, err
			}
			if c.Reference == a.Reference {
				return c, nil
			}
		}
		if len(list.Data) < 20 {
			return domain.Charge{}, domain.ErrNotFound
		}
	}
	return domain.Charge{}, fmt.Errorf("charge search page limit reached")
}
func (o *Omise) request(ctx context.Context, method, path string, form url.Values) ([]byte, error) {
	if !o.Ready() {
		return nil, domain.Conflict("PAYMENT_UNAVAILABLE")
	}
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, o.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(o.SecretKey, "")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	response, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			o.authFailureUntil.Store(time.Now().Add(domain.ProviderAuthRetryAfterSeconds * time.Second).UnixNano())
		}
		return nil, &domain.ProviderHTTPError{StatusCode: response.StatusCode}
	}
	o.authFailureUntil.Store(0)
	raw, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > 2*1024*1024 {
		return nil, fmt.Errorf("provider response too large")
	}
	return raw, nil
}
func decodeCharge(raw []byte) (domain.Charge, error) {
	var p struct {
		ID             string          `json:"id"`
		Reference      string          `json:"ref_id"`
		Status         string          `json:"status"`
		Amount         int64           `json:"amount"`
		Currency       string          `json:"currency"`
		Live           bool            `json:"livemode"`
		Expired        bool            `json:"expired"`
		ExpiresAt      string          `json:"expires_at"`
		AuthorizeURI   string          `json:"authorize_uri"`
		ReturnURI      string          `json:"return_uri"`
		FailureCode    string          `json:"failure_code"`
		FailureMessage string          `json:"failure_message"`
		Transaction    json.RawMessage `json:"transaction"`
		Metadata       struct {
			Attempt string `json:"payment_attempt_id"`
			Request string `json:"request_no"`
		} `json:"metadata"`
		Source struct {
			ID        string `json:"id"`
			Scannable struct {
				Image struct {
					DownloadURI string `json:"download_uri"`
				} `json:"image"`
			} `json:"scannable_code"`
		} `json:"source"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return domain.Charge{}, err
	}
	var expiry time.Time
	if p.ExpiresAt != "" {
		var err error
		expiry, err = time.Parse(time.RFC3339, p.ExpiresAt)
		if err != nil {
			return domain.Charge{}, err
		}
	}
	var transaction string
	if len(p.Transaction) > 0 && string(p.Transaction) != "null" {
		if err := json.Unmarshal(p.Transaction, &transaction); err != nil {
			var expanded struct {
				ID string `json:"id"`
			}
			if err = json.Unmarshal(p.Transaction, &expanded); err != nil {
				return domain.Charge{}, err
			}
			transaction = expanded.ID
		}
	}
	if p.Reference == "" {
		p.Reference = p.Metadata.Attempt
	}
	return domain.Charge{ID: p.ID, Reference: p.Reference, RequestNo: p.Metadata.Request, Status: p.Status, Amount: p.Amount, Currency: strings.ToUpper(p.Currency), Live: p.Live, Expired: p.Expired, ExpiresAt: expiry, SourceID: p.Source.ID, TransactionID: transaction, AuthorizeURI: p.AuthorizeURI, ReturnURI: p.ReturnURI, QRCodeURI: p.Source.Scannable.Image.DownloadURI, FailureCode: p.FailureCode, FailureMessage: p.FailureMessage, Raw: string(raw)}, nil
}
func (o *Omise) VerifySignature(raw, signature, timestamp string) bool {
	// Authenticity is also verified by retrieving the charge with the server key.
	if o.WebhookSecret == "" {
		return true
	}
	if signature == "" || timestamp == "" {
		return false
	}
	key, err := base64.StdEncoding.DecodeString(o.WebhookSecret)
	if err != nil {
		key = []byte(o.WebhookSecret)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(timestamp + "." + raw))
	expected := hex.EncodeToString(mac.Sum(nil))
	for _, candidate := range strings.Split(signature, ",") {
		if hmac.Equal([]byte(expected), []byte(strings.TrimSpace(candidate))) {
			return true
		}
	}
	return false
}
