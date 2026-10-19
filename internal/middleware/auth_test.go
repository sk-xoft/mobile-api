package middleware

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthenticateRejectsRevokedExpiredAndDisabledSessions(t *testing.T) {
	for _, kind := range []string{"revoked", "expired-session", "disabled-user", "legacy-session"} {
		t.Run(kind, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			rows := sqlmock.NewRows([]string{"USER_ID", "IS_ONLINE"})
			switch kind {
			case "revoked":
				rows.AddRow("user", "NO")
			case "disabled-user":
				rows.AddRow("user", "YES")
			}
			mock.ExpectQuery("SELECT s.USER_ID, s.IS_ONLINE").WithArgs("session").WillReturnRows(rows)
			if kind == "disabled-user" {
				mock.ExpectQuery("SELECT USERNAME.*USER_STATUS='A'").WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"USERNAME"}))
			}
			app := fiber.New()
			app.Get("/", Authenticate("secret", db), func(c *fiber.Ctx) error { return c.SendStatus(200) })
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "smart-seaman-mobile-api", Subject: "crew@example.com", ID: "session", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}).SignedString([]byte("secret"))
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 401 {
				t.Fatalf("status=%d", resp.StatusCode)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAuthenticateAllowsActiveSecureSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT s.USER_ID, s.IS_ONLINE").WithArgs("session").WillReturnRows(sqlmock.NewRows([]string{"USER_ID", "IS_ONLINE"}).AddRow("user", "YES"))
	mock.ExpectQuery("SELECT USERNAME.*USER_STATUS='A'").WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"USERNAME"}).AddRow("crew@example.com"))
	app := fiber.New()
	app.Get("/", Authenticate("secret", db), func(c *fiber.Ctx) error {
		if c.Locals("user.mobile_uuid") != "user" {
			t.Error("missing authenticated identity")
		}
		return c.SendStatus(200)
	})
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "smart-seaman-mobile-api", Subject: "crew@example.com", ID: "session", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}).SignedString([]byte("secret"))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
