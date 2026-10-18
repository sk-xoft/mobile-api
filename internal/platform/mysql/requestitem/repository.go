package requestitemmysql

import (
	"database/sql"
	domain "github.com/smartseaman/mobile-api/internal/domain/requestitem"
)

type Repository = domain.Repository

func NewRepository(db *sql.DB) *Repository { return domain.NewRepository(db) }
