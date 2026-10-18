package requestitem

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type Service struct {
	repo    repositoryPort
	storage Storage
	path    string
}

type repositoryPort interface {
	ProfileItems(context.Context, string) ([]Item, error)
	Files(context.Context, string, string, string) ([]File, error)
	UpsertProfile(context.Context, string, string, string, string, string, string, string, int64) (string, error)
	Renewal(context.Context, string, string, string) (string, string, string, string, error)
	UpsertRenewal(context.Context, string, string, string, string, string, string, string, int64) (string, error)
}

func NewService(r repositoryPort, s Storage, path string) *Service {
	return &Service{repo: r, storage: s, path: strings.Trim(path, "/")}
}
func (s *Service) List(ctx context.Context, user string) ([]Item, error) {
	return s.repo.ProfileItems(ctx, user)
}
func (s *Service) Preview(ctx context.Context, user, item string, renewalID string) (PreviewResponse, error) {
	files, e := s.repo.Files(ctx, user, item, renewalID)
	return PreviewResponse{ItemCode: item, DocumentType: firstType(files), Files: files}, e
}
func (s *Service) Upload(ctx context.Context, user, item, kind, slot, name string, content []byte, renewalNo string) (UploadResponse, error) {
	item = strings.ToUpper(strings.TrimSpace(item))
	kind = strings.ToUpper(strings.TrimSpace(kind))
	slot = strings.ToUpper(strings.TrimSpace(slot))
	if e := valid(item, kind, slot); e != nil {
		return UploadResponse{}, e
	}
	mime, e := detect(content)
	if e != nil {
		return UploadResponse{}, e
	}
	if renewalNo == "" {
		key := s.key(user, uuid.NewString())
		if e = s.storage.Put(ctx, key, mime, content); e != nil {
			return UploadResponse{}, fmt.Errorf("upload profile request item to storage: %w", e)
		}
		id, e := s.repo.UpsertProfile(ctx, user, item, kind, slot, key, name, mime, int64(len(content)))
		if e != nil {
			return UploadResponse{}, fmt.Errorf("save profile request item: %w", e)
		}
		files, e := s.repo.Files(ctx, user, item, "")
		return UploadResponse{ProfileID: id, ItemCode: item, StorageScope: "PROFILE", DocumentType: kind, Complete: complete(kind, files), Files: files}, e
	}
	requestID, itemID, requestStatus, approveStatus, e := s.repo.Renewal(ctx, user, renewalNo, item)
	if e != nil {
		return UploadResponse{}, e
	}
	if requestStatus != "PAYMENT_PENDING" && requestStatus != "PENDING_APPLICANT_CORRECTION" {
		return UploadResponse{}, &BusinessError{Code: "MA00005", Detail: "documentStatus"}
	}
	if requestStatus == "PENDING_APPLICANT_CORRECTION" && approveStatus != "FIX" {
		return UploadResponse{}, &BusinessError{Code: "MA00005", Detail: "documentRequestItemStatus"}
	}
	key := s.key(itemID, uuid.NewString())
	if e = s.storage.Put(ctx, key, mime, content); e != nil {
		return UploadResponse{}, fmt.Errorf("upload renewal request item to storage: %w", e)
	}
	id, e := s.repo.UpsertRenewal(ctx, itemID, item, kind, slot, key, name, mime, int64(len(content)))
	if e != nil {
		return UploadResponse{}, fmt.Errorf("save renewal request item: %w", e)
	}
	files, e := s.repo.Files(ctx, user, item, requestID)
	return UploadResponse{RequestID: requestID, RequestNo: renewalNo, RequestFileID: id, ItemCode: item, StorageScope: "REQUEST", DocumentType: kind, Complete: complete(kind, files), Files: files}, e
}

func (s *Service) key(scope, fileID string) string {
	if strings.Contains(s.path, "%s") {
		return fmt.Sprintf(s.path, scope, fileID)
	}
	return s.path + "/" + scope + "/" + fileID
}

func valid(item, kind, slot string) error {
	if item == "MRI001" {
		if !((kind == "ID_CARD" && (slot == "MAIN" || slot == "FRONT" || slot == "BACK")) || (kind == "PASSPORT" && slot == "MAIN")) {
			return &BusinessError{Code: "MA00005", Detail: "documentType/slotCode"}
		}
		return nil
	}
	if kind != "GENERAL" || slot != "MAIN" {
		return &BusinessError{Code: "MA00005", Detail: "documentType/slotCode"}
	}
	return nil
}
func detect(b []byte) (string, error) {
	if len(b) >= 8 && string(b[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png", nil
	}
	if len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff {
		return "image/jpeg", nil
	}
	if len(b) >= 5 && string(b[:5]) == "%PDF-" {
		return "application/pdf", nil
	}
	return "", &BusinessError{Code: "MA00005", Detail: "fileMimeType"}
}
func complete(kind string, files []File) bool {
	if kind == "ID_CARD" {
		front, back, main := false, false, false
		for _, f := range files {
			if f.UploadedAt != "" {
				front = front || f.SlotCode == "FRONT"
				back = back || f.SlotCode == "BACK"
				main = main || f.SlotCode == "MAIN"
			}
		}
		return main || (front && back)
	}
	return len(files) > 0
}
func firstType(files []File) string {
	if len(files) > 0 {
		return files[0].DocumentType
	}
	return ""
}

type BusinessError struct{ Code, Detail string }

func (e *BusinessError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Detail) }
