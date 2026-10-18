package eventlogmysql

import (
	"database/sql"

	domain "github.com/smartseaman/mobile-api/internal/domain/eventlog"
)

type Repository = domain.MySQLRepository

func NewRepository(db *sql.DB) *Repository {
	return domain.NewMySQLRepository(db)
}
