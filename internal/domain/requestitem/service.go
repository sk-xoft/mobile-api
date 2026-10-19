package requestitem

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smartseaman/mobile-api/internal/constants"
)

type Service struct {
	repo    repositoryPort
	storage Storage
	path    string
}

type repositoryPort interface {
	ValidateItem(context.Context, string, string) error
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
	items, err := s.repo.ProfileItems(ctx, user)
	if err != nil {
		return nil, err
	}
	for i := range items {
		files, err := s.repo.Files(ctx, user, items[i].ItemCode, "")
		if err != nil {
			return nil, err
		}
		files = currentFiles(files)
		items[i].Files, items[i].DocumentType = files, firstType(files)
		items[i].FileUploaded = 0
		items[i].FilePath, items[i].FileUploadedAt, items[i].CheckResult, items[i].CheckNote = "", "", "", ""
		if len(files) > 0 {
			items[i].FileUploaded = 1
			f := files[0]
			items[i].ProfileID, items[i].FilePath, items[i].FileUploadedAt = f.FileID, f.FilePath, f.UploadedAt
			items[i].CheckResult, items[i].CheckNote = f.CheckResult, f.CheckNote
		}
		items[i].DocumentStatus = "MISSING"
		if complete(items[i].DocumentType, files) {
			items[i].DocumentStatus = "COMPLETE"
		}
	}
	return items, nil
}
func (s *Service) Preview(ctx context.Context, user, item string, renewalID string) (PreviewResponse, error) {
	files, e := s.repo.Files(ctx, user, item, renewalID)
	files = currentFiles(files)
	return PreviewResponse{ItemCode: item, DocumentType: firstType(files), Files: files}, e
}
func (s *Service) Upload(ctx context.Context, user, item, kind, slot, name string, content []byte, renewalNo string) (UploadResponse, error) {
	item = strings.ToUpper(strings.TrimSpace(item))
	kind = strings.ToUpper(strings.TrimSpace(kind))
	slot = strings.ToUpper(strings.TrimSpace(slot))
	if len(content) == 0 || len(content) > 10*1024*1024 {
		return UploadResponse{}, &BusinessError{Code: constants.CodeInvalidInput, Detail: "file"}
	}
	if e := valid(item, kind, slot); e != nil {
		return UploadResponse{}, e
	}
	mime, e := detect(content)
	if e != nil {
		return UploadResponse{}, e
	}
	scope := ""
	if renewalNo == "" {
		scope = constants.DocumentStorageScopeProfile
	}
	if e = s.repo.ValidateItem(ctx, item, scope); e != nil {
		return UploadResponse{}, e
	}
	if renewalNo == "" {
		key := s.key(user, uuid.NewString())
		if e = s.storage.Put(ctx, key, mime, content); e != nil {
			return UploadResponse{}, fmt.Errorf("upload profile request item to storage: %w", e)
		}
		id, e := s.repo.UpsertProfile(ctx, user, item, kind, slot, key, name, mime, int64(len(content)))
		if e != nil {
			return UploadResponse{}, s.cleanup(ctx, key, fmt.Errorf("save profile request item: %w", e))
		}
		files, e := s.repo.Files(ctx, user, item, "")
		files = currentFiles(files)
		return UploadResponse{ProfileID: id, ItemCode: item, StorageScope: constants.DocumentStorageScopeProfile, DocumentType: kind, Complete: complete(kind, files), Files: files}, e
	}
	requestID, itemID, requestStatus, approveStatus, e := s.repo.Renewal(ctx, user, renewalNo, item)
	if e != nil {
		return UploadResponse{}, e
	}
	if requestStatus != constants.StatusPaymentPending && requestStatus != constants.StatusApplicantCorrection {
		return UploadResponse{}, &BusinessError{Code: constants.CodeInvalidInput, Detail: "documentStatus"}
	}
	if requestStatus == constants.StatusApplicantCorrection && approveStatus != constants.RequestItemApprovalFix {
		return UploadResponse{}, &BusinessError{Code: constants.CodeInvalidInput, Detail: "documentRequestItemStatus"}
	}
	key := s.key(itemID, uuid.NewString())
	if e = s.storage.Put(ctx, key, mime, content); e != nil {
		return UploadResponse{}, fmt.Errorf("upload renewal request item to storage: %w", e)
	}
	id, e := s.repo.UpsertRenewal(ctx, itemID, item, kind, slot, key, name, mime, int64(len(content)))
	if e != nil {
		return UploadResponse{}, s.cleanup(ctx, key, fmt.Errorf("save renewal request item: %w", e))
	}
	files, e := s.repo.Files(ctx, user, item, requestID)
	files = currentFiles(files)
	return UploadResponse{RequestID: requestID, RequestNo: renewalNo, RequestFileID: id, ItemCode: item, StorageScope: constants.DocumentStorageScopeRequest, DocumentType: kind, Complete: complete(kind, files), Files: files}, e
}

func (s *Service) cleanup(ctx context.Context, key string, cause error) error {
	if errors.Is(cause, ErrSaveUncertain) {
		return cause
	}
	if storage, ok := s.storage.(interface {
		Delete(context.Context, string) error
	}); ok {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if err := storage.Delete(cleanupCtx, key); err != nil {
			return errors.Join(cause, fmt.Errorf("cleanup uploaded request item: %w", err))
		}
	}
	return cause
}

// Keep older document types as history; the most recently uploaded type is active.
func currentFiles(files []File) []File {
	if len(files) == 0 {
		return []File{}
	}
	kind := files[0].DocumentType
	latest := ""
	for _, f := range files {
		if f.UploadedAt > latest {
			latest, kind = f.UploadedAt, f.DocumentType
		}
	}
	out := make([]File, 0, len(files))
	for _, f := range files {
		if f.DocumentType == kind {
			out = append(out, f)
		}
	}
	return out
}

func (s *Service) key(scope, fileID string) string {
	if strings.Contains(s.path, "%s") {
		return fmt.Sprintf(s.path, scope, fileID)
	}
	return s.path + "/" + scope + "/" + fileID
}

func valid(item, kind, slot string) error {
	if item == "MRI001" {
		if !((kind == constants.DocumentTypeIDCard && (slot == constants.DocumentSlotMain || slot == constants.DocumentSlotFront || slot == constants.DocumentSlotBack)) || (kind == constants.DocumentTypePassport && slot == constants.DocumentSlotMain)) {
			return &BusinessError{Code: constants.CodeInvalidInput, Detail: "documentType/slotCode"}
		}
		return nil
	}
	if kind != constants.DocumentTypeGeneral || slot != constants.DocumentSlotMain {
		return &BusinessError{Code: constants.CodeInvalidInput, Detail: "documentType/slotCode"}
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
	return "", &BusinessError{Code: constants.CodeInvalidInput, Detail: "fileMimeType"}
}
func complete(kind string, files []File) bool {
	if kind == constants.DocumentTypeIDCard {
		front, back, main := false, false, false
		for _, f := range files {
			if f.DocumentType == kind && f.UploadedAt != "" {
				front = front || f.SlotCode == constants.DocumentSlotFront
				back = back || f.SlotCode == constants.DocumentSlotBack
				main = main || f.SlotCode == constants.DocumentSlotMain
			}
		}
		return main || (front && back)
	}
	for _, f := range files {
		if f.DocumentType == kind && f.SlotCode == constants.DocumentSlotMain && f.UploadedAt != "" {
			return true
		}
	}
	return false
}
func firstType(files []File) string {
	if len(files) > 0 {
		return files[0].DocumentType
	}
	return ""
}

type BusinessError struct{ Code, Detail string }

func (e *BusinessError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Detail) }
