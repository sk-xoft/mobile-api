package authmysql

import (
	"database/sql"
	domain "github.com/smartseaman/mobile-api/internal/domain/auth"
)

type Repository struct {
	*domain.Repository
	db             *sql.DB
	successMessage successMessageCache
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{Repository: domain.NewRepository(db), db: db}
}
