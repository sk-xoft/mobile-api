package requestitem

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("document item not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) active(ctx context.Context, item string) (bool, error) {
	var n int
	e := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_document_master_request_item WHERE document_master_items_code=? AND is_active='YES'", item).Scan(&n)
	return n == 1, e
}
func (r *Repository) ProfileItems(ctx context.Context, user string) ([]Item, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT m.id,m.document_master_items_code,m.storage_scope,m.sort_order,COALESCE(p.file_uploaded,0),COALESCE(p.file_path,''),COALESCE(p.file_uploaded_at,''),COALESCE(p.check_result,''),COALESCE(p.check_note,'') FROM m_document_master_request_item m LEFT JOIN (SELECT document_master_request_item_code,MIN(id) id FROM m_document_profile_request_item WHERE mobile_user_uuid=? GROUP BY document_master_request_item_code) x ON x.document_master_request_item_code=m.document_master_items_code LEFT JOIN m_document_profile_request_item p ON p.id=x.id WHERE m.is_active='YES' AND m.storage_scope='PROFILE' ORDER BY m.sort_order,m.document_master_items_code`, user)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]Item, 0)
	for rows.Next() {
		var i Item
		if e := rows.Scan(&i.ID, &i.ItemCode, &i.StorageScope, &i.SortOrder, &i.FileUploaded, &i.FilePath, &i.FileUploadedAt, &i.CheckResult, &i.CheckNote); e != nil {
			return nil, e
		}
		i.DocumentStatus = "MISSING"
		if i.FileUploaded > 0 {
			i.DocumentStatus = "COMPLETE"
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
func (r *Repository) Files(ctx context.Context, user, item string, renewalID string) ([]File, error) {
	query := `SELECT id,document_type,slot_code,COALESCE(file_path,''),COALESCE(original_file_name,''),COALESCE(mime_type,''),COALESCE(file_size,0),COALESCE(file_uploaded_at,''),COALESCE(check_result,''),COALESCE(check_note,''),COALESCE(is_updated,0) FROM m_document_profile_request_item WHERE mobile_user_uuid=? AND document_master_request_item_code=?`
	args := []any{user, item}
	if renewalID != "" {
		query = `SELECT f.id,f.document_type,f.slot_code,COALESCE(f.file_path,''),COALESCE(f.original_file_name,''),COALESCE(f.mime_type,''),COALESCE(f.file_size,0),COALESCE(f.file_uploaded_at,''),COALESCE(f.check_result,''),COALESCE(f.check_note,''),COALESCE(f.is_updated,0) FROM m_document_request_item_files f INNER JOIN m_document_request_items i ON i.id=f.request_item_id INNER JOIN m_document_request r ON r.id=i.request_id WHERE r.id=? AND r.mobile_user_uuid=? AND i.document_master_request_item_code=?`
		args = []any{renewalID, user, item}
	}
	rows, e := r.db.QueryContext(ctx, query, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]File, 0)
	for rows.Next() {
		var f File
		if e := rows.Scan(&f.FileID, &f.DocumentType, &f.SlotCode, &f.FilePath, &f.OriginalName, &f.MimeType, &f.FileSize, &f.UploadedAt, &f.CheckResult, &f.CheckNote, &f.IsUpdated); e != nil {
			return nil, e
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
func (r *Repository) UpsertProfile(ctx context.Context, user, item, kind, slot, key, name, mime string, size int64) (string, error) {
	var id string
	e := r.db.QueryRowContext(ctx, "SELECT id FROM m_document_profile_request_item WHERE mobile_user_uuid=? AND document_master_request_item_code=? AND document_type=? AND slot_code=?", user, item, kind, slot).Scan(&id)
	if errors.Is(e, sql.ErrNoRows) {
		id = uuid.NewString()
		_, e = r.db.ExecContext(ctx, `INSERT INTO m_document_profile_request_item(id,mobile_user_uuid,document_master_request_item_code,document_type,slot_code,file_uploaded,file_path,original_file_name,mime_type,file_size,file_uploaded_at) SELECT ?,?,?,?, ?,1,?,?,?,?,NOW() FROM m_document_master_request_item WHERE document_master_items_code=? AND is_active='YES'`, id, user, item, kind, slot, key, name, mime, size, item)
		return id, e
	}
	if e != nil {
		return "", e
	}
	_, e = r.db.ExecContext(ctx, "UPDATE m_document_profile_request_item SET file_uploaded=1,file_path=?,original_file_name=?,mime_type=?,file_size=?,file_uploaded_at=NOW(),check_result=NULL,check_note=NULL,is_updated=1 WHERE id=?", key, name, mime, size, id)
	return id, e
}
func (r *Repository) Renewal(ctx context.Context, user, no, item string) (string, string, string, string, error) {
	var id, requestID, requestStatus, approveStatus string
	e := r.db.QueryRowContext(ctx, "SELECT r.id,i.id,s.document_status_code,COALESCE(i.approve_status,'') FROM m_document_request r INNER JOIN m_document_request_items i ON i.request_id=r.id INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.request_no=? AND r.mobile_user_uuid=? AND r.is_active='YES' AND i.document_master_request_item_code=?", no, user, item).Scan(&requestID, &id, &requestStatus, &approveStatus)
	if errors.Is(e, sql.ErrNoRows) {
		return "", "", "", "", ErrNotFound
	}
	return requestID, id, requestStatus, approveStatus, e
}
func (r *Repository) UpsertRenewal(ctx context.Context, itemID, item, kind, slot, key, name, mime string, size int64) (string, error) {
	var id string
	e := r.db.QueryRowContext(ctx, "SELECT id FROM m_document_request_item_files WHERE request_item_id=? AND document_type=? AND slot_code=?", itemID, kind, slot).Scan(&id)
	if errors.Is(e, sql.ErrNoRows) {
		id = uuid.NewString()
		_, e = r.db.ExecContext(ctx, `INSERT INTO m_document_request_item_files(id,request_item_id,document_master_request_item_code,document_type,slot_code,file_uploaded,file_path,original_file_name,mime_type,file_size,file_uploaded_at) VALUES(?,?,?,?,?,1,?,?,?,?,NOW())`, id, itemID, item, kind, slot, key, name, mime, size)
		return id, e
	}
	if e != nil {
		return "", e
	}
	_, e = r.db.ExecContext(ctx, "UPDATE m_document_request_item_files SET file_uploaded=1,file_path=?,original_file_name=?,mime_type=?,file_size=?,file_uploaded_at=NOW(),check_result=NULL,check_note=NULL,is_updated=1 WHERE id=?", key, name, mime, size, id)
	return id, e
}
