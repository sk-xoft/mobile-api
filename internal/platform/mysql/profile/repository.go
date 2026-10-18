package profilemysql

import (
	"database/sql"
	domain "github.com/smartseaman/mobile-api/internal/domain/profile"
)

type Repository = domain.Repository

func NewRepository(db *sql.DB) *Repository { return domain.NewRepository(db) }
