package document

import (
	"context"
	"errors"
	"testing"
)

type fakeDocumentRepository struct {
	user            User
	certificate     Certificate
	certificateErr  error
	pageItems       []Item
	pageTotal       int
	pageErr         error
	closeItems      []Item
	closeTotal      int
	closeErr        error
	createErr       error
	updateErr       error
	deleteErr       error
	createdFileName string
	updatedFileName string
	updatedUsername string
}

func (f *fakeDocumentRepository) FindUser(context.Context, string) (User, error) {
	return f.user, nil
}
func (f *fakeDocumentRepository) Certificate(context.Context, string, string) (Certificate, error) {
	return f.certificate, f.certificateErr
}
func (f *fakeDocumentRepository) CreateCertificate(_ context.Context, _, username string, _ CreateRequest, fileName string) error {
	f.updatedUsername, f.createdFileName = username, fileName
	return f.createErr
}
func (f *fakeDocumentRepository) UpdateCertificate(_ context.Context, _, username string, _ UpdateRequest, fileName string) error {
	f.updatedUsername, f.updatedFileName = username, fileName
	return f.updateErr
}
func (f *fakeDocumentRepository) DeleteCertificate(context.Context, string, string) error {
	return f.deleteErr
}
func (f *fakeDocumentRepository) Page(context.Context, string, string, int) ([]Item, int, error) {
	return f.pageItems, f.pageTotal, f.pageErr
}
func (f *fakeDocumentRepository) CloseToExpiration(context.Context, string, int) ([]Item, int, error) {
	return f.closeItems, f.closeTotal, f.closeErr
}

type fakeDocumentStorage struct {
	putKey, putType string
	putContent      []byte
	putErr          error
	getContent      []byte
	getType         string
	getErr          error
}

func (f *fakeDocumentStorage) Put(_ context.Context, key, contentType string, content []byte) error {
	f.putKey, f.putType, f.putContent = key, contentType, content
	return f.putErr
}
func (f *fakeDocumentStorage) Get(context.Context, string) ([]byte, string, error) {
	return f.getContent, f.getType, f.getErr
}

func validPNGDataURL() string {
	return "data:image/png;base64,iVBORw0KGgo="
}

func TestServicePageNormalizesDatesAndPagination(t *testing.T) {
	repository := &fakeDocumentRepository{
		pageItems: []Item{{CertStartDate: "2025-01-01", CertEndDate: "2025-12-31"}},
		pageTotal: 2,
	}
	data, err := NewService(repository, &fakeDocumentStorage{}, "documents/%s/certs/%s").Page(context.Background(), "user-1", "COT", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !data.Last || data.ItemTotal != 2 || len(data.Items) != 1 {
		t.Fatalf("unexpected page: %#v", data)
	}
	if data.Items[0].DisYear == "" || data.Items[0].CertExpiredFlag != "Y" {
		t.Fatalf("dates were not normalized: %#v", data.Items[0])
	}
}

func TestServicePageRejectsNegativeOffsetAndPropagatesRepositoryError(t *testing.T) {
	service := NewService(&fakeDocumentRepository{pageErr: errors.New("database down")}, &fakeDocumentStorage{}, "")
	if _, err := service.Page(context.Background(), "user-1", "COT", -1); err == nil {
		t.Fatal("negative offset should fail")
	}
	if _, err := service.Page(context.Background(), "user-1", "COT", 0); err == nil || err.Error() != "database down" {
		t.Fatalf("repository error = %v", err)
	}
}

func TestServiceCloseToExpirationUsesSamePageRules(t *testing.T) {
	repository := &fakeDocumentRepository{
		closeItems: []Item{{CertEndDate: "9999-99-99"}},
		closeTotal: 1,
	}
	data, err := NewService(repository, &fakeDocumentStorage{}, "").CloseToExpiration(context.Background(), "user-1", 0)
	if err != nil || !data.Last || data.Items[0].CertExpiredFlag != "N" {
		t.Fatalf("close-to-expiration result = %#v, %v", data, err)
	}
}

func TestServiceCreateStoresDecodedFileAndPersistsName(t *testing.T) {
	repository := &fakeDocumentRepository{user: User{Username: "crew"}, certificateErr: ErrNotFound}
	storage := &fakeDocumentStorage{}
	request := CreateRequest{DocumentCode: "DOC001", CertStartDate: "2026-01-01", CertEndDate: "9999-99-99", FileCert: validPNGDataURL(), FileCertName: "cert.png"}
	response, err := NewService(repository, storage, "documents/%s/certs/%s").Create(context.Background(), "user-1", request)
	if err != nil {
		t.Fatal(err)
	}
	if response.DocumentCode != request.DocumentCode || repository.updatedUsername != "crew" || repository.createdFileName == "" {
		t.Fatalf("unexpected create result: %#v repo=%#v", response, repository)
	}
	if storage.putType != "image/png" || storage.putKey == "" || len(storage.putContent) == 0 {
		t.Fatalf("file was not stored: %#v", storage)
	}
}

func TestServiceCreateRejectsDuplicateAndStorageFailure(t *testing.T) {
	duplicate := &fakeDocumentRepository{user: User{Username: "crew"}, certificateErr: nil}
	request := CreateRequest{DocumentCode: "DOC001", CertStartDate: "2026-01-01", FileCert: validPNGDataURL()}
	_, err := NewService(duplicate, &fakeDocumentStorage{}, "").Create(context.Background(), "user-1", request)
	var business *BusinessError
	if !errors.As(err, &business) || business.Code != "MA00017" {
		t.Fatalf("duplicate error = %v", err)
	}

	repository := &fakeDocumentRepository{user: User{Username: "crew"}, certificateErr: ErrNotFound}
	_, err = NewService(repository, &fakeDocumentStorage{putErr: errors.New("storage down")}, "").Create(context.Background(), "user-1", request)
	if err == nil || err.Error() != "storage down" {
		t.Fatalf("storage error = %v", err)
	}
}

func TestServiceUpdateWithAndWithoutFile(t *testing.T) {
	request := UpdateRequest{DocumentCode: "DOC001", CertStartDate: "2026-01-01", IsChangeFile: "Y", FileCert: validPNGDataURL(), FileCertName: "new.png"}
	repository := &fakeDocumentRepository{user: User{Username: "crew"}}
	storage := &fakeDocumentStorage{}
	if _, err := NewService(repository, storage, "documents/%s/certs/%s").Update(context.Background(), "user-1", request); err != nil {
		t.Fatal(err)
	}
	if repository.updatedFileName == "" || storage.putType != "image/png" {
		t.Fatalf("changed file was not persisted: repo=%#v storage=%#v", repository, storage)
	}

	repository = &fakeDocumentRepository{user: User{Username: "crew"}}
	request.IsChangeFile = "N"
	if _, err := NewService(repository, &fakeDocumentStorage{}, "").Update(context.Background(), "user-1", request); err != nil {
		t.Fatal(err)
	}
	if repository.updatedFileName != "" {
		t.Fatal("file should not be replaced when IsChangeFile is N")
	}
}

func TestServiceEditViewAndDelete(t *testing.T) {
	repository := &fakeDocumentRepository{certificate: Certificate{DocumentCode: "DOC001", StartDate: "2026-01-01", EndDate: "", FileName: "stored.png", OriginalFileName: "original.png"}}
	storage := &fakeDocumentStorage{getContent: []byte("file"), getType: "image/png"}
	service := NewService(repository, storage, "documents/%s/certs/%s")
	edit, err := service.Edit(context.Background(), "user-1", "DOC001")
	if err != nil || edit.CertEndDateType != "N" || edit.CertEndDate != "9999-99-99" {
		t.Fatalf("edit = %#v, %v", edit, err)
	}
	content, contentType, err := service.View(context.Background(), "user-1", "DOC001")
	if err != nil || string(content) != "file" || contentType != "image/png" {
		t.Fatalf("view = %q, %q, %v", content, contentType, err)
	}
	if result, err := service.Delete(context.Background(), "user-1", "DOC001"); err != nil || result != "success" {
		t.Fatalf("delete = %q, %v", result, err)
	}
}

func TestServiceViewRejectsMissingFileAndDeletePropagatesError(t *testing.T) {
	repository := &fakeDocumentRepository{certificate: Certificate{DocumentCode: "DOC001"}}
	service := NewService(repository, &fakeDocumentStorage{}, "")
	if _, _, err := service.View(context.Background(), "user-1", "DOC001"); err == nil {
		t.Fatal("missing file should fail")
	}
	repository.deleteErr = errors.New("delete failed")
	if _, err := service.Delete(context.Background(), "user-1", "DOC001"); err == nil {
		t.Fatal("delete error should propagate")
	}
}

func TestServiceKeySupportsTemplateAndPrefix(t *testing.T) {
	if got := (&Service{pathTemplate: "documents/%s/certs/%s"}).key("u", "f"); got != "documents/u/certs/f" {
		t.Fatalf("template key = %q", got)
	}
	if got := (&Service{pathTemplate: "/documents"}).key("u", "f"); got != "documents/u/f" {
		t.Fatalf("prefix key = %q", got)
	}
}
