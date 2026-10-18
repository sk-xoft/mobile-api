package contentmysql

import (
	"database/sql"
	domain "github.com/smartseaman/mobile-api/internal/domain/content"
)

type Repository = domain.Repository

func NewRepository(db *sql.DB) *Repository { return domain.NewRepository(db) }
