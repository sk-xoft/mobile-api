package profile

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
)

type storedImage struct {
	content []byte
	err     error
}

func (s storedImage) Put(context.Context, string, string, []byte) error { return nil }
func (s storedImage) Get(context.Context, string) ([]byte, string, error) {
	return s.content, "text/plain", s.err
}

func TestProfileImageServesLegacyAndBinary(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n\x00")
	for _, legacy := range []bool{true, false} {
		t.Run(map[bool]string{true: "legacy", false: "binary"}[legacy], func(t *testing.T) {
			db, m, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			m.ExpectQuery(`SELECT MOBILE_UUID,.*FROM m_mobile_users WHERE MOBILE_UUID = \? LIMIT 1`).WithArgs("user-1").WillReturnRows(sqlmock.NewRows([]string{"uuid", "smart", "username", "email", "first", "last", "mobile", "company", "position", "picture", "status", "dob"}).AddRow("user-1", "1", "user", "email", "First", "Last", "0800000000", "C", "P", "file", "A", "2000-01-01"))
			body := png
			if legacy {
				body = []byte(base64.StdEncoding.EncodeToString(png))
			}
			handler := NewHandler(NewService(NewRepository(db), storedImage{content: body}, "profiles"))
			app := fiber.New()
			app.Get("/image", func(c *fiber.Ctx) error { c.Locals("user.mobile_uuid", "user-1"); return handler.Image(c) })
			res, err := app.Test(httptest.NewRequest("GET", "/image", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			got, _ := io.ReadAll(res.Body)
			if res.StatusCode != 200 || res.Header.Get("Content-Type") != "image/png" || !bytes.Equal(got, png) {
				t.Fatalf("status=%d type=%s length=%d", res.StatusCode, res.Header.Get("Content-Type"), len(got))
			}
			if err := m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestProfileImagePropagatesStorageError(t *testing.T) {
	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m.ExpectQuery(`SELECT MOBILE_UUID,.*FROM m_mobile_users`).WillReturnRows(sqlmock.NewRows([]string{"uuid", "smart", "username", "email", "first", "last", "mobile", "company", "position", "picture", "status", "dob"}).AddRow("user-1", "1", "user", "email", "First", "Last", "0800000000", "C", "P", "file", "A", "2000-01-01"))
	failure := errors.New("storage failed")
	_, _, err = NewService(NewRepository(db), storedImage{err: failure}, "profiles").Image(context.Background(), "user-1")
	if !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
