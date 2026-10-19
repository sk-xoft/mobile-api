package payment_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	domain "github.com/smartseaman/mobile-api/internal/domain/payment"
	provider "github.com/smartseaman/mobile-api/internal/platform/payment"
)

// Memory storage isolates orchestration; SQL transaction behavior has separate tests.
type memoryStore struct {
	mu           sync.Mutex
	request      domain.Request
	attempts     map[string]domain.Attempt
	keys         map[string]string
	active, paid string
	review       bool
	transitions  int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{request: domain.Request{ID: "request", No: "REQ001", UserID: "owner", Status: "PAYMENT_PENDING", Amount: 18000, Active: true}, attempts: map[string]domain.Attempt{}, keys: map[string]string{}}
}
func (m *memoryStore) Request(_ context.Context, id, user string) (domain.Request, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id != m.request.ID || user != m.request.UserID {
		return domain.Request{}, domain.ErrNotFound
	}
	return m.request, nil
}
func (m *memoryStore) read(id string) (domain.Attempt, error) {
	a, ok := m.attempts[id]
	if !ok {
		return a, domain.ErrNotFound
	}
	a.Review = a.Review || m.review || !m.request.Active
	a.Paid = m.paid != ""
	return a, nil
}
func (m *memoryStore) Get(_ context.Context, id string) (domain.Attempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.read(id)
}
func (m *memoryStore) ByKey(_ context.Context, id, key string) (domain.Attempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.read(m.keys[key])
}
func (m *memoryStore) Current(_ context.Context, id string) (domain.Attempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.paid != "" {
		return m.read(m.paid)
	}
	if m.active != "" {
		return m.read(m.active)
	}
	var latest domain.Attempt
	for _, a := range m.attempts {
		if a.CreatedAt.After(latest.CreatedAt) {
			latest = a
		}
	}
	if latest.ID == "" {
		return latest, domain.ErrNotFound
	}
	return m.read(latest.ID)
}
func (m *memoryStore) Reserve(_ context.Context, a domain.Attempt) (domain.Attempt, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id, ok := m.keys[a.Key]; ok {
		old, _ := m.read(id)
		if old.Fingerprint != a.Fingerprint {
			return a, false, domain.Conflict("PAYMENT_IDEMPOTENCY_CONFLICT")
		}
		return old, false, nil
	}
	if m.review {
		return a, false, domain.Conflict("PAYMENT_RECONCILIATION_REQUIRED")
	}
	if m.paid != "" {
		return a, false, domain.Conflict("PAYMENT_ALREADY_PAID")
	}
	if m.active != "" {
		return a, false, domain.Conflict("PAYMENT_IN_PROGRESS")
	}
	if !m.request.Active || m.request.Status != "PAYMENT_PENDING" {
		return a, false, domain.Conflict("PAYMENT_REQUEST_NOT_PAYABLE")
	}
	m.attempts[a.ID] = a
	m.keys[a.Key] = a.ID
	m.active = a.ID
	return a, true, nil
}
func (m *memoryStore) Acquire(_ context.Context, id, owner string, _ bool) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.attempts[id]
	if a.Lease != "" {
		return false, nil
	}
	a.Lease = owner
	m.attempts[id] = a
	return true, nil
}
func (m *memoryStore) MarkSent(_ context.Context, a domain.Attempt) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.attempts[a.ID]
	if current.Creation != "RESERVED" || !m.request.Active || m.paid != "" {
		return false, nil
	}
	current.Creation = "SENT"
	current.SentAt = time.Now()
	m.attempts[a.ID] = current
	return true, nil
}
func (m *memoryStore) Apply(_ context.Context, a domain.Attempt, c domain.Charge) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.attempts[a.ID]
	status := domain.Status(c)
	if current.Status == "SUCCESS" {
		status = "SUCCESS"
	}
	if status == "SUCCESS" {
		if m.paid != "" && m.paid != a.ID {
			m.review = true
		} else if m.paid == "" {
			m.paid = a.ID
			if m.request.Active && m.request.Status == "PAYMENT_PENDING" {
				m.transitions++
				m.request.Status = "PENDING_DOCUMENT_REVIEW"
			} else {
				m.review = true
			}
		}
	}
	current.Status = status
	current.Charge = c
	current.Creation = "CONFIRMED"
	current.Lease = ""
	m.attempts[a.ID] = current
	if status == "EXPIRED" || status == "FAILED" || status == "SUCCESS" {
		if m.active == a.ID {
			m.active = ""
		}
	}
	return nil
}
func (m *memoryStore) Retry(_ context.Context, a domain.Attempt, review bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a = m.attempts[a.ID]
	a.Lease = ""
	a.Review = a.Review || review
	m.review = m.review || review
	m.attempts[a.ID] = a
	return nil
}
func (m *memoryStore) Due(context.Context, int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := []string{}
	for id := range m.attempts {
		ids = append(ids, id)
	}
	return ids, nil
}
func (m *memoryStore) Find(_ context.Context, c domain.Charge) (domain.Attempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.attempts {
		if a.Reference == c.Reference || a.Charge.ID == c.ID {
			return m.read(a.ID)
		}
	}
	return domain.Attempt{}, domain.ErrNotFound
}

type fakeOmise struct {
	mu                sync.Mutex
	charges           map[string]map[string]any
	posts             int
	postStatus        int
	loseResponse      bool
	expiredAtCreation bool
	wrongAmount       bool
}

func newFakeOmise() *fakeOmise { return &fakeOmise{charges: map[string]map[string]any{}} }
func (f *fakeOmise) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	user, _, ok := r.BasicAuth()
	if !ok || user != "test-key" {
		w.WriteHeader(401)
		return
	}
	if r.Method == "POST" {
		f.posts++
		if f.postStatus != 0 {
			w.WriteHeader(f.postStatus)
			fmt.Fprint(w, `{"error":"failure"}`)
			return
		}
		_ = r.ParseForm()
		ref := r.Form.Get("ref_id")
		id := "chrg_" + ref
		expiry := time.Now().Add(time.Minute)
		if f.expiredAtCreation {
			expiry = time.Now().Add(-time.Minute)
		}
		amount := int64(18000)
		if f.wrongAmount {
			amount = 1
		}
		c := map[string]any{"id": id, "ref_id": ref, "status": "pending", "amount": amount, "currency": "THB", "livemode": false, "expires_at": expiry.UTC().Format(time.RFC3339), "metadata": map[string]string{"request_no": "REQ001", "payment_attempt_id": ref}, "source": map[string]any{"id": "src_1", "scannable_code": map[string]any{"image": map[string]string{"download_uri": "https://example.test/qr/" + ref}}}, "transaction": "trxn_1"}
		f.charges[id] = c
		if f.loseResponse { // Charge exists, but the transport fails before any response.
			connection, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				connection.Close()
			}
			return
		}
		_ = json.NewEncoder(w).Encode(c)
		return
	}
	if r.URL.Path == "/charges" {
		data := []map[string]any{}
		for _, c := range f.charges {
			data = append(data, c)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
		return
	}
	c, ok := f.charges[strings.TrimPrefix(r.URL.Path, "/charges/")]
	if !ok {
		w.WriteHeader(404)
		return
	}
	_ = json.NewEncoder(w).Encode(c)
}
func (f *fakeOmise) set(id, status string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.charges[id]["status"] = status
	f.charges[id]["expired"] = status == "expired"
}
func (f *fakeOmise) count() int { f.mu.Lock(); defer f.mu.Unlock(); return f.posts }
func setup(t *testing.T, f *fakeOmise) (*domain.Service, *memoryStore) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(server.Close)
	store := newMemoryStore()
	client := provider.NewOmise(server.URL, "test-key", "", 15*time.Minute)
	return domain.NewService(store, client, false), store
}
func input(key string) domain.Input {
	return domain.Input{PaymentMethod: "promptpay", IdempotencyKey: key}
}
func assertCode(t *testing.T, err error, code string) {
	t.Helper()
	var e *domain.Error
	if !errors.As(err, &e) || e.Code != code {
		t.Fatalf("got %v, want %s", err, code)
	}
}

func TestRealHTTPExpiredQRReplacementAndReplay(t *testing.T) {
	f := newFakeOmise()
	service, _ := setup(t, f)
	ctx := context.Background()
	first, err := service.Create(ctx, "owner", "request", input("one"))
	if err != nil {
		t.Fatal(err)
	}
	if first.NextAction != "WAIT_FOR_PAYMENT" || first.QRCodeURI == "" {
		t.Fatalf("first=%+v", first)
	}
	replay, err := service.Create(ctx, "owner", "request", input("one"))
	if err != nil || replay.TransactionID != first.TransactionID || replay.QRCodeURI != first.QRCodeURI || f.count() != 1 {
		t.Fatalf("replay=%+v, %v", replay, err)
	}
	_, err = service.Create(ctx, "owner", "request", input("different"))
	assertCode(t, err, "PAYMENT_IN_PROGRESS")
	f.set(first.ProviderChargeID, "expired")
	expired, err := service.Get(ctx, "owner", "request", first.TransactionID)
	if err != nil || expired.NextAction != "CREATE_NEW_QR" || expired.QRCodeURI != "" {
		t.Fatalf("expired=%+v, %v", expired, err)
	}
	same, err := service.Create(ctx, "owner", "request", input("one"))
	if err != nil || same.TransactionID != first.TransactionID || f.count() != 1 {
		t.Fatalf("old key created charge: %+v %v", same, err)
	}
	replacement, err := service.Create(ctx, "owner", "request", input("two"))
	if err != nil || replacement.TransactionID == first.TransactionID || f.count() != 2 {
		t.Fatalf("replacement=%+v, %v", replacement, err)
	}
}
func TestRealHTTPClockExpiryCannotAuthorizeNewCharge(t *testing.T) {
	f := newFakeOmise()
	f.expiredAtCreation = true
	service, _ := setup(t, f)
	p, err := service.Create(context.Background(), "owner", "request", input("one"))
	if err != nil {
		t.Fatal(err)
	}
	if !p.QRExpired || p.NextAction != "CHECK_PAYMENT" || p.Status != "PROCESSING" || p.QRCodeURI != "" {
		t.Fatalf("payment=%+v", p)
	}
	_, err = service.Create(context.Background(), "owner", "request", input("two"))
	assertCode(t, err, "PAYMENT_IN_PROGRESS")
	if f.count() != 1 {
		t.Fatal("created another charge")
	}
}
func TestRealHTTPLostCreateResponseRecoveredWithoutPostingAgain(t *testing.T) {
	f := newFakeOmise()
	f.loseResponse = true
	service, _ := setup(t, f)
	p, err := service.Create(context.Background(), "owner", "request", input("one"))
	if err != nil || p.Status != "PROCESSING" {
		t.Fatalf("payment=%+v,error=%v", p, err)
	}
	recovered, err := service.Create(context.Background(), "owner", "request", input("one"))
	if err != nil || recovered.ProviderChargeID == "" || recovered.TransactionID != p.TransactionID || f.count() != 1 {
		t.Fatalf("recovered=%+v,error=%v", recovered, err)
	}
}
func TestRealHTTPLateSuccessAndDuplicateWebhook(t *testing.T) {
	f := newFakeOmise()
	service, store := setup(t, f)
	ctx := context.Background()
	old, err := service.Create(ctx, "owner", "request", input("one"))
	if err != nil {
		t.Fatal(err)
	}
	f.set(old.ProviderChargeID, "expired")
	_, err = service.Get(ctx, "owner", "request", old.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	replacement, err := service.Create(ctx, "owner", "request", input("two"))
	if err != nil {
		t.Fatal(err)
	}
	f.set(old.ProviderChargeID, "successful")
	event := fmt.Sprintf(`{"key":"charge.complete","data":{"id":%q}}`, old.ProviderChargeID)
	for i := 0; i < 2; i++ {
		if err = service.Webhook(ctx, event, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	if store.transitions != 1 {
		t.Fatalf("transitions=%d", store.transitions)
	}
	current, err := service.Current(ctx, "owner", "request")
	if err != nil || current.TransactionID != old.TransactionID || current.NextAction != "COMPLETE" {
		t.Fatalf("current=%+v,error=%v", current, err)
	}
	_, err = service.Create(ctx, "owner", "request", input("three"))
	assertCode(t, err, "PAYMENT_ALREADY_PAID")
	f.set(replacement.ProviderChargeID, "successful")
	event = fmt.Sprintf(`{"key":"charge.complete","data":{"id":%q}}`, replacement.ProviderChargeID)
	if err = service.Webhook(ctx, event, "", ""); err != nil {
		t.Fatal(err)
	}
	current, err = service.Current(ctx, "owner", "request")
	if err != nil || current.NextAction != "CONTACT_SUPPORT" || store.transitions != 1 {
		t.Fatalf("duplicate=%+v,error=%v", current, err)
	}
}
func TestRealHTTPConcurrentCreationOwnerAndFingerprint(t *testing.T) {
	f := newFakeOmise()
	service, _ := setup(t, f)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := service.Create(ctx, "owner", "request", input("same")); err != nil {
				t.Errorf("concurrent create: %v", err)
			}
		}()
	}
	wg.Wait()
	if f.count() != 1 {
		t.Fatalf("POSTs=%d", f.count())
	}
	different := input("same")
	different.ReturnURI = "app://changed"
	_, err := service.Create(ctx, "owner", "request", different)
	assertCode(t, err, "PAYMENT_IDEMPOTENCY_CONFLICT")
	if _, err = service.Current(ctx, "attacker", "request"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("owner bypass")
	}
}
func TestRealHTTPProviderFailuresRemainUncertain(t *testing.T) {
	for _, code := range []int{409, 429, 500} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			f := newFakeOmise()
			f.postStatus = code
			service, _ := setup(t, f)
			p, err := service.Create(context.Background(), "owner", "request", input("one"))
			if err != nil || p.NextAction != "CHECK_PAYMENT" {
				t.Fatalf("payment=%+v,error=%v", p, err)
			}
			_, err = service.Create(context.Background(), "owner", "request", input("two"))
			assertCode(t, err, "PAYMENT_IN_PROGRESS")
			if f.count() != 1 {
				t.Fatal("uncertain charge retried as POST")
			}
		})
	}
}
func TestRealHTTPMismatchRequiresReviewAndCancelledSuccessCannotFulfil(t *testing.T) {
	f := newFakeOmise()
	f.wrongAmount = true
	service, _ := setup(t, f)
	p, err := service.Create(context.Background(), "owner", "request", input("one"))
	if err != nil || p.NextAction != "CONTACT_SUPPORT" {
		t.Fatalf("mismatch=%+v,error=%v", p, err)
	}
	f = newFakeOmise()
	service, store := setup(t, f)
	p, err = service.Create(context.Background(), "owner", "request", input("one"))
	if err != nil {
		t.Fatal(err)
	}
	store.request.Active = false
	f.set(p.ProviderChargeID, "successful")
	event := fmt.Sprintf(`{"key":"charge.complete","data":{"id":%q}}`, p.ProviderChargeID)
	if err = service.Webhook(context.Background(), event, "", ""); err != nil {
		t.Fatal(err)
	}
	if store.transitions != 0 || !store.review {
		t.Fatal("cancelled request fulfilled")
	}
}

func TestWorkersReconcileAfterAppClosedAndStopCleanly(t *testing.T) {
	f := newFakeOmise()
	service, store := setup(t, f)
	p, err := service.Create(context.Background(), "owner", "request", input("one"))
	if err != nil {
		t.Fatal(err)
	}
	f.set(p.ProviderChargeID, "successful")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 2; i++ {
		workers.Add(1)
		go func() { defer workers.Done(); service.Run(ctx) }()
	}
	go func() { workers.Wait(); close(done) }()
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout.C:
			t.Fatal("worker did not reconcile")
		case <-ticker.C:
			store.mu.Lock()
			count := store.transitions
			store.mu.Unlock()
			if count == 1 {
				cancel()
				select {
				case <-done:
					return
				case <-time.After(time.Second):
					t.Fatal("worker did not stop")
				}
			}
		}
	}
}
