package requestitem

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeRequestItemRepo struct {
	validateErr                        error
	items                              []Item
	files                              []File
	err                                error
	renewalID, itemID, status, approve string
	profileID, requestFileID           string
	lastScope, lastItem, lastKind      string
}

func (f *fakeRequestItemRepo) ValidateItem(context.Context, string, string) error {
	return f.validateErr
}

func (f *fakeRequestItemRepo) ProfileItems(context.Context, string) ([]Item, error) {
	return f.items, f.err
}
func (f *fakeRequestItemRepo) Files(context.Context, string, string, string) ([]File, error) {
	return f.files, f.err
}
func (f *fakeRequestItemRepo) UpsertProfile(_ context.Context, _, item, kind, _, _, _, _ string, _ int64) (string, error) {
	f.lastItem, f.lastKind, f.lastScope = item, kind, "PROFILE"
	return f.profileID, f.err
}
func (f *fakeRequestItemRepo) Renewal(context.Context, string, string, string) (string, string, string, string, error) {
	return f.renewalID, f.itemID, f.status, f.approve, f.err
}
func (f *fakeRequestItemRepo) UpsertRenewal(_ context.Context, _, item, kind, _, _, _, _ string, _ int64) (string, error) {
	f.lastItem, f.lastKind, f.lastScope = item, kind, "REQUEST"
	return f.requestFileID, f.err
}

type fakeRequestItemStorage struct {
	err  error
	key  string
	mime string
	body []byte
}

func (f *fakeRequestItemStorage) Put(_ context.Context, key, mime string, body []byte) error {
	f.key, f.mime, f.body = key, mime, body
	return f.err
}
func (f *fakeRequestItemStorage) Get(context.Context, string) ([]byte, string, error) {
	return nil, "", nil
}

func TestRequestItemHelpers(t *testing.T) {
	t.Run("validates document combinations", func(t *testing.T) {
		for _, tc := range []struct {
			item, kind, slot string
			valid            bool
		}{
			{"MRI001", "ID_CARD", "FRONT", true}, {"MRI001", "ID_CARD", "BACK", true},
			{"MRI001", "PASSPORT", "MAIN", true}, {"MRI001", "GENERAL", "MAIN", false},
			{"OTHER", "GENERAL", "MAIN", true}, {"OTHER", "ID_CARD", "MAIN", false},
		} {
			if got := valid(tc.item, tc.kind, tc.slot) == nil; got != tc.valid {
				t.Errorf("valid(%q,%q,%q)=%v, want %v", tc.item, tc.kind, tc.slot, got, tc.valid)
			}
		}
	})
	t.Run("detects supported mime types", func(t *testing.T) {
		cases := []struct {
			body []byte
			mime string
		}{
			{[]byte("\x89PNG\r\n\x1a\nrest"), "image/png"},
			{[]byte{0xff, 0xd8, 0xff, 0x00}, "image/jpeg"},
			{[]byte("%PDF-1.7"), "application/pdf"},
		}
		for _, tc := range cases {
			if got, err := detect(tc.body); err != nil || got != tc.mime {
				t.Errorf("detect()=(%q,%v)", got, err)
			}
		}
		if _, err := detect([]byte("bad")); err == nil {
			t.Fatal("invalid mime should fail")
		}
	})
	if !complete("ID_CARD", []File{{DocumentType: "ID_CARD", SlotCode: "FRONT", UploadedAt: "now"}, {DocumentType: "ID_CARD", SlotCode: "BACK", UploadedAt: "now"}}) {
		t.Error("front and back should complete ID card")
	}
	if complete("ID_CARD", []File{{SlotCode: "FRONT", UploadedAt: "now"}}) {
		t.Error("a single ID card side should not complete the ID card")
	}
	if complete("ID_CARD", []File{{SlotCode: "FRONT"}, {SlotCode: "BACK"}}) {
		t.Error("files without upload time should not complete ID card")
	}
	if firstType([]File{{DocumentType: "PDF"}}) != "PDF" || firstType(nil) != "" {
		t.Error("unexpected firstType result")
	}
}

func TestServiceListAndPreview(t *testing.T) {
	repo := &fakeRequestItemRepo{items: []Item{{ItemCode: "A"}}, files: []File{{DocumentType: "PDF"}}}
	svc := NewService(repo, &fakeRequestItemStorage{}, "uploads")
	items, err := svc.List(context.Background(), "user")
	if err != nil || len(items) != 1 {
		t.Fatalf("List()=(%v,%v)", items, err)
	}
	preview, err := svc.Preview(context.Background(), "user", "A", "")
	if err != nil || preview.DocumentType != "PDF" {
		t.Fatalf("Preview()=(%+v,%v)", preview, err)
	}
	repo.err = errors.New("files failed")
	if _, err = svc.Preview(context.Background(), "user", "A", ""); err == nil {
		t.Fatal("Preview should propagate repository errors")
	}
}

func TestServiceUploadProfile(t *testing.T) {
	repo := &fakeRequestItemRepo{profileID: "profile-1", files: []File{{DocumentType: "GENERAL", SlotCode: "MAIN", UploadedAt: "now"}}}
	storage := &fakeRequestItemStorage{}
	svc := NewService(repo, storage, "uploads/%s/%s")
	resp, err := svc.Upload(context.Background(), "user", " other ", " general ", " main ", "a.pdf", []byte("%PDF-1.7"), "")
	if err != nil || resp.ProfileID != "profile-1" || resp.ItemCode != "OTHER" || resp.DocumentType != "GENERAL" || !resp.Complete {
		t.Fatalf("Upload()=(%+v,%v)", resp, err)
	}
	if !strings.HasPrefix(storage.key, "uploads/user/") || storage.mime != "application/pdf" {
		t.Errorf("storage=%q %q", storage.key, storage.mime)
	}

	for _, tc := range []struct {
		name                string
		repoErr, storageErr error
		want                string
	}{
		{"storage error", nil, errors.New("put"), "upload profile request item to storage"},
		{"repository error", errors.New("save"), nil, "save profile request item"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeRequestItemRepo{err: tc.repoErr, files: repo.files}
			st := &fakeRequestItemStorage{err: tc.storageErr}
			_, err := NewService(r, st, "uploads").Upload(context.Background(), "u", "OTHER", "GENERAL", "MAIN", "x.pdf", []byte("%PDF-1.7"), "")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
		})
	}
	if _, err = svc.Upload(context.Background(), "u", "OTHER", "BAD", "MAIN", "x", []byte("%PDF-1.7"), ""); err == nil {
		t.Fatal("invalid document combination should fail")
	}
}

func TestServiceUploadRenewal(t *testing.T) {
	base := func(status, approve string) (*fakeRequestItemRepo, *fakeRequestItemStorage) {
		return &fakeRequestItemRepo{renewalID: "request-1", itemID: "item-1", status: status, approve: approve, requestFileID: "file-1", files: []File{{SlotCode: "MAIN", UploadedAt: "now"}}}, &fakeRequestItemStorage{}
	}
	for _, tc := range []struct{ name, status, approve string }{
		{"payment pending", "PAYMENT_PENDING", ""}, {"correction", "PENDING_APPLICANT_CORRECTION", "FIX"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, st := base(tc.status, tc.approve)
			resp, err := NewService(r, st, "uploads").Upload(context.Background(), "u", "OTHER", "GENERAL", "MAIN", "x.pdf", []byte("%PDF-1.7"), "RN-1")
			if err != nil || resp.RequestID != "request-1" || resp.RequestFileID != "file-1" || resp.StorageScope != "REQUEST" {
				t.Fatalf("Upload()=(%+v,%v)", resp, err)
			}
		})
	}
	for _, tc := range []struct{ status, approve, detail string }{
		{"APPROVED", "", "documentStatus"}, {"PENDING_APPLICANT_CORRECTION", "PASS", "documentRequestItemStatus"},
	} {
		r, st := base(tc.status, tc.approve)
		_, err := NewService(r, st, "uploads").Upload(context.Background(), "u", "OTHER", "GENERAL", "MAIN", "x.pdf", []byte("%PDF-1.7"), "RN-1")
		if err == nil || !strings.Contains(err.Error(), tc.detail) {
			t.Errorf("error=%v, want %q", err, tc.detail)
		}
	}
	r, st := base("PAYMENT_PENDING", "")
	r.err = errors.New("lookup")
	if _, err := NewService(r, st, "uploads").Upload(context.Background(), "u", "OTHER", "GENERAL", "MAIN", "x.pdf", []byte("%PDF-1.7"), "RN-1"); err == nil {
		t.Fatal("renewal lookup error should propagate")
	}
}
