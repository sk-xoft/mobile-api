package app

import (
	"strings"
	"testing"

	"github.com/smartseaman/mobile-api/internal/config"
)

func TestNewRejectsMissingDatabaseBeforeServingRequests(t *testing.T) {
	for name, dsn := range map[string]string{"empty": "", "whitespace": " \t\n"} {
		t.Run(name, func(t *testing.T) {
			application, err := New(config.Config{Database: config.Database{DSN: dsn}})
			if application != nil {
				defer application.Close()
				t.Fatal("application initialized without a database; requests could panic")
			}
			if err == nil || !strings.Contains(err.Error(), "DATABASE_DSN is required") {
				t.Fatalf("expected missing database configuration error, got %v", err)
			}
		})
	}
}
