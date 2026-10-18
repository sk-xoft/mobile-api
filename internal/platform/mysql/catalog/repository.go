package catalogmysql

import (
	"database/sql"
	domain "github.com/smartseaman/mobile-api/internal/domain/catalog"
)

type Repository = domain.Repository

func NewRepository(db *sql.DB) *Repository { return domain.NewRepository(db) }
