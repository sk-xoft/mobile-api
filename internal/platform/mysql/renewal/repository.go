package renewalmysql

import (
	"database/sql"
	domain "github.com/smartseaman/mobile-api/internal/domain/renewal"
)

type Repository = domain.Repository

func NewRepository(db *sql.DB, omise *domain.OmiseClient) *Repository {
	return domain.NewRepository(db, omise)
}
