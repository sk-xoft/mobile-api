package auth

import (
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
