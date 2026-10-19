package authmysql

import (
	"context"
	"database/sql/driver"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func expectSuccessMessage(mock sqlmock.Sqlmock, english, thai string) {
	mock.ExpectQuery("SELECT MESSAGE_DESCRIPTION_EN, MESSAGE_DESCRIPTION_TH").WithArgs("MA00000").WillReturnRows(sqlmock.NewRows([]string{"en", "th"}).AddRow(english, thai))
}

func TestSuccessMessageCachesBothLanguagesAndExpires(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := NewRepository(db)
	expectSuccessMessage(mock, "Success", "สำเร็จ")
	for _, language := range []string{"en", "EN", "th", "TH", ""} {
		want := "สำเร็จ"
		if language == "en" || language == "EN" {
			want = "Success"
		}
		if got := r.Message(context.Background(), "MA00000", language); got != want {
			t.Fatalf("%s: got %q", language, got)
		}
	}
	r.successMessage.expiresAt = time.Now().Add(-time.Second)
	expectSuccessMessage(mock, "Updated success", "ข้อความใหม่")
	if got := r.Message(context.Background(), "MA00000", "en"); got != "Updated success" {
		t.Fatal(got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSuccessMessageCoalescesConcurrentColdRequests(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := NewRepository(db)
	expectSuccessMessage(mock, "Success", "สำเร็จ")
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := r.Message(context.Background(), "MA00000", "en"); got != "Success" {
				t.Errorf("got %q", got)
			}
		}()
	}
	wg.Wait()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMessageDoesNotCacheFailuresOrErrorDescriptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := NewRepository(db)
	mock.ExpectQuery("SELECT MESSAGE_DESCRIPTION_EN").WithArgs("MA00000").WillReturnError(errors.New("unavailable"))
	if got := r.Message(context.Background(), "MA00000", "en"); got != "Message code : MA00000 is not matching in master." {
		t.Fatal(got)
	}
	expectSuccessMessage(mock, "Success", "สำเร็จ")
	if got := r.Message(context.Background(), "MA00000", "en"); got != "Success" {
		t.Fatal(got)
	}
	for _, value := range []string{"First", "Changed"} {
		mock.ExpectQuery("SELECT MESSAGE_DESCRIPTION_EN").WithArgs("MA00003").WillReturnRows(sqlmock.NewRows([]string{"en", "th"}).AddRow(value, nil))
		if got := r.Message(context.Background(), "MA00003", "en"); got != value {
			t.Fatal(got)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkSuccessMessage(b *testing.B) {
	for _, cached := range []bool{false, true} {
		name := "Uncached"
		if cached {
			name = "WarmCache"
		}
		b.Run(name, func(b *testing.B) {
			db, mock, err := sqlmock.New()
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()
			r := NewRepository(db)
			if cached {
				expectSuccessMessage(mock, "Success", "สำเร็จ")
				r.Message(context.Background(), "MA00000", "en")
			} else {
				for i := 0; i < b.N; i++ {
					mock.ExpectQuery("SELECT MESSAGE_DESCRIPTION_EN").WithArgs("MA00000").WillReturnRows(sqlmock.NewRows([]string{"en", "th"}).AddRow(driver.Value("Success"), driver.Value("สำเร็จ")))
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if cached {
					r.Message(context.Background(), "MA00000", "en")
				} else {
					r.Repository.Message(context.Background(), "MA00000", "en")
				}
			}
			b.StopTimer()
			if err := mock.ExpectationsWereMet(); err != nil {
				b.Fatal(err)
			}
		})
	}
}
