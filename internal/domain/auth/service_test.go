package auth

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueTokenMatchesLegacyClaimsAndLifetime(t *testing.T) {
	fixed := time.Date(2026, time.October, 17, 10, 30, 0, 0, time.FixedZone("Asia/Bangkok", 7*60*60))
	service := &Service{secret: []byte("01234567890123456789012345678901"), now: func() time.Time { return fixed }}

	signed, issuedAt, err := service.issueToken("crew@example.com", "session-1")
	if err != nil {
		t.Fatalf("issueToken() error = %v", err)
	}
	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(signed, claims, func(*jwt.Token) (any, error) { return service.secret, nil }, jwt.WithoutClaimsValidation())
	if err != nil || !parsed.Valid {
		t.Fatalf("parse token: %v", err)
	}
	if claims.Subject != "crew@example.com" || claims.ID != "session-1" || claims.Issuer != "smart-seaman-mobile-api" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if issuedAt != fixed || claims.ExpiresAt.Sub(claims.IssuedAt.Time) != 30*time.Minute {
		t.Fatalf("unexpected token timestamps: %#v", claims)
	}
}

type securityRepository struct {
	repositoryPort
	user                                      User
	actionHash, purpose, password             string
	sessionID, refreshHash, oldHash, nextHash string
	consumed                                  bool
}

func (r *securityRepository) ConsumeActionToken(_ context.Context, hash, purpose, password string) (User, error) {
	if r.consumed || hash != r.actionHash || purpose != r.purpose {
		return User{}, ErrNotFound
	}
	r.consumed = true
	r.password = password
	return r.user, nil
}
func (r *securityRepository) FindUserByUsername(context.Context, string) (User, error) {
	return r.user, nil
}
func (r *securityRepository) FindUserByUUID(context.Context, string) (User, error) {
	return r.user, nil
}
func (r *securityRepository) CreateSecureSession(_ context.Context, id, user, token, hash, correlation string, now time.Time) error {
	r.sessionID = id
	r.refreshHash = hash
	return nil
}
func (r *securityRepository) FindRefreshSession(_ context.Context, hash string) (string, string, error) {
	if hash != r.refreshHash {
		return "", "", ErrNotFound
	}
	return r.sessionID, r.user.MobileUUID, nil
}
func (r *securityRepository) RotateRefreshSession(_ context.Context, id, old, next, token string) error {
	if old != r.refreshHash {
		return ErrNotFound
	}
	r.oldHash = old
	r.nextHash = next
	r.refreshHash = next
	return nil
}
func TestResetRequiresSingleUsePurposeBoundToken(t *testing.T) {
	r := &securityRepository{user: User{Email: "crew@example.com"}, actionHash: tokenHash("emailed-secret"), purpose: "reset"}
	s := NewService(r, nil, "secret")
	for _, uid := range []string{"user-uuid", "wrong-token"} {
		if _, err := s.CompletePasswordReset(context.Background(), ForgotPasswordRequest{UID: uid, Password: "new-password", ConfirmPassword: "new-password"}); err == nil {
			t.Fatal("accepted identifier instead of reset token")
		}
	}
	if _, err := s.Activate(context.Background(), "emailed-secret"); err == nil {
		t.Fatal("reset token activated account")
	}
	req := ForgotPasswordRequest{UID: "emailed-secret", Password: "new-password", ConfirmPassword: "new-password"}
	if _, err := s.CompletePasswordReset(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(r.password), []byte(req.Password)) != nil {
		t.Fatal("password was not hashed")
	}
	if _, err := s.CompletePasswordReset(context.Background(), req); err == nil {
		t.Fatal("reset token reused")
	}
}
func TestLoginSeparatesRefreshFromJWTAndRotates(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	r := &securityRepository{user: User{MobileUUID: "user", Username: "crew@example.com", Password: string(hash), Status: "A"}}
	s := NewService(r, nil, "01234567890123456789012345678901")
	login, err := s.Login(context.Background(), LoginRequest{Username: r.user.Username, Password: "password"}, "correlation")
	if err != nil {
		t.Fatal(err)
	}
	claims := &jwt.RegisteredClaims{}
	if _, err := jwt.ParseWithClaims(login.Token, claims, func(*jwt.Token) (any, error) { return s.secret, nil }); err != nil {
		t.Fatal(err)
	}
	if claims.ID == login.RefreshToken || claims.ID != r.sessionID || r.refreshHash != tokenHash(login.RefreshToken) {
		t.Fatal("refresh secret exposed in session identifier or stored in plaintext")
	}
	if _, err := s.Refresh(context.Background(), claims.ID); err == nil {
		t.Fatal("session id used as refresh secret")
	}
	refreshed, err := s.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.RefreshToken == login.RefreshToken || refreshed.RefreshToken == "" || r.nextHash != tokenHash(refreshed.RefreshToken) {
		t.Fatal("refresh token not rotated")
	}
	if _, err := s.Refresh(context.Background(), login.RefreshToken); err == nil {
		t.Fatal("old refresh token reused")
	}
	r.user.Status = "D"
	if _, err := s.Refresh(context.Background(), refreshed.RefreshToken); err == nil {
		t.Fatal("disabled account refreshed")
	}
}

// This isolates login CPU work; it intentionally excludes network/database time.
func BenchmarkLoginCost8(b *testing.B) {
	hash, err := bcrypt.GenerateFromPassword([]byte("benchmark-password"), 8)
	if err != nil {
		b.Fatal(err)
	}
	r := &securityRepository{user: User{MobileUUID: "user", Username: "crew@example.com", Password: string(hash), Status: "A"}}
	s := NewService(r, nil, "01234567890123456789012345678901")
	request := LoginRequest{Username: r.user.Username, Password: "benchmark-password"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Login(context.Background(), request, "benchmark"); err != nil {
			b.Fatal(err)
		}
	}
}
