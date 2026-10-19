package renewal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type omiseClient struct {
	baseURL, secretKey, webhookSecret string
	httpClient                        *http.Client
	qrExpire                          time.Duration
}

type OmiseClient = omiseClient
type omiseCharge struct {
	ID             string
	Status         string
	SourceID       string
	TransactionID  string
	AuthorizeURI   string
	ReturnURI      string
	FailureCode    string
	FailureMessage string
	QRCodeURI      string
	ExpiresAt      time.Time
	Livemode       bool
	Raw            string
}

func newOmiseClient(qrExpire ...time.Duration) *omiseClient {
	expire := 15 * time.Minute
	if len(qrExpire) > 0 && qrExpire[0] > 0 {
		expire = qrExpire[0]
	}
	return &omiseClient{
		baseURL:       strings.TrimRight(envOr("OMISE_API_URL", "https://api.omise.co"), "/"),
		secretKey:     os.Getenv("OMISE_SECRET_KEY"),
		webhookSecret: os.Getenv("OMISE_WEBHOOK_SECRET"),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		qrExpire: expire,
	}
}

func NewOmiseClient(qrExpire ...time.Duration) *omiseClient { return newOmiseClient(qrExpire...) }
func (c *omiseClient) enabled() bool                        { return strings.TrimSpace(c.secretKey) != "" }
func (c *omiseClient) CreateCharge(amount float64, requestNo string, input PaymentRequest) (omiseCharge, error) {
	if !c.enabled() {
		return omiseCharge{}, fmt.Errorf("OMISE_SECRET_KEY is not configured")
	}
	form := url.Values{}
	form.Set("amount", fmt.Sprintf("%.0f", amount*100))
	form.Set("currency", "THB")
	form.Set("description", "Document renewal "+requestNo)
	form.Set("metadata[request_no]", requestNo)
	form.Set("metadata[idempotency_key]", input.IdempotencyKey)
	form.Set("source[type]", input.PaymentMethod)
	if input.PlatformType != "" {
		form.Set("source[platform_type]", input.PlatformType)
	}
	if input.ReturnURI != "" {
		form.Set("return_uri", input.ReturnURI)
	}
	if strings.EqualFold(input.PaymentMethod, "promptpay") {
		form.Set("expires_at", time.Now().UTC().Add(c.qrExpire).Format(time.RFC3339))
	}
	return c.request(http.MethodPost, "/charges", form, input.IdempotencyKey)
}
func (c *omiseClient) RetrieveCharge(id string) (omiseCharge, error) {
	if !c.enabled() {
		return omiseCharge{}, fmt.Errorf("OMISE_SECRET_KEY is not configured")
	}
	return c.request(http.MethodGet, "/charges/"+url.PathEscape(id), nil, "")
}
func (c *omiseClient) request(method, path string, form url.Values, idempotency string) (omiseCharge, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, e := http.NewRequest(method, c.baseURL+path, body)
	if e != nil {
		return omiseCharge{}, e
	}
	req.SetBasicAuth(c.secretKey, "")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if idempotency != "" {
		req.Header.Set("Idempotency-Key", idempotency)
	}
	resp, e := c.httpClient.Do(req)
	if e != nil {
		return omiseCharge{}, e
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return omiseCharge{}, fmt.Errorf("omise http %d: %s", resp.StatusCode, string(raw))
	}
	var payload struct {
		ID             string `json:"id"`
		Status         string `json:"status"`
		AuthorizeURI   string `json:"authorize_uri"`
		ReturnURI      string `json:"return_uri"`
		FailureCode    string `json:"failure_code"`
		FailureMessage string `json:"failure_message"`
		ExpiresAt      string `json:"expires_at"`
		Livemode       bool   `json:"livemode"`
		Transaction    struct {
			ID string `json:"id"`
		} `json:"transaction"`
		Source struct {
			ID            string `json:"id"`
			ScannableCode struct {
				Image struct {
					DownloadURI string `json:"download_uri"`
				} `json:"image"`
			} `json:"scannable_code"`
		} `json:"source"`
	}
	if e = json.Unmarshal(raw, &payload); e != nil {
		return omiseCharge{}, e
	}
	expiresAt := time.Time{}
	if payload.ExpiresAt != "" {
		expiresAt, e = time.Parse(time.RFC3339, payload.ExpiresAt)
		if e != nil {
			return omiseCharge{}, fmt.Errorf("parse omise expires_at: %w", e)
		}
	}
	return omiseCharge{
			ID:             payload.ID,
			Status:         payload.Status,
			SourceID:       payload.Source.ID,
			TransactionID:  payload.Transaction.ID,
			AuthorizeURI:   payload.AuthorizeURI,
			ReturnURI:      payload.ReturnURI,
			FailureCode:    payload.FailureCode,
			FailureMessage: payload.FailureMessage,
			QRCodeURI:      payload.Source.ScannableCode.Image.DownloadURI,
			ExpiresAt:      expiresAt,
			Livemode:       payload.Livemode,
			Raw:            string(raw),
		},
		nil
}
func (c *omiseClient) VerifySignature(raw, signature, timestamp string) bool {
	if c.webhookSecret == "" {
		return true
	}
	if signature == "" || timestamp == "" {
		return false
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(c.webhookSecret))
	if err != nil {
		key = []byte(c.webhookSecret)
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
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
