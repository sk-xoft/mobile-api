package eventlog

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Create(ctx context.Context, event Event) error {
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO event_logs
		(id,event_name,actor_user_id,aggregate_type,aggregate_id,request_id,trace_id,metadata,created_at)
		VALUES(?,?,?,?,?,?,?,?,NOW())`,
		uuid.NewString(), nullString(event.Name), nullString(event.ActorUserID), nullString(event.AggregateType),
		nullString(event.AggregateID), nullString(event.RequestID), nullString(event.TraceID), metadata)
	return err
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
