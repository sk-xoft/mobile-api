package requestitem

import "context"

type Storage interface {
	Put(context.Context, string, string, []byte) error
	Get(context.Context, string) ([]byte, string, error)
}
type File struct {
	FileID       string `json:"fileId"`
	DocumentType string `json:"documentType"`
	SlotCode     string `json:"slotCode"`
	FilePath     string `json:"filePath,omitempty"`
	OriginalName string `json:"originalFileName"`
	MimeType     string `json:"mimeType"`
	FileSize     int64  `json:"fileSize"`
	UploadedAt   string `json:"fileUploadedAt,omitempty"`
	CheckResult  string `json:"checkResult,omitempty"`
	CheckNote    string `json:"checkNote,omitempty"`
	IsUpdated    bool   `json:"isUpdated,omitempty"`
}
type Item struct {
	ID             string `json:"id"`
	ProfileID      string `json:"profileRequestItemId,omitempty"`
	DocumentCode   string `json:"documentCode,omitempty"`
	ItemCode       string `json:"documentMasterRequestItemCode"`
	StorageScope   string `json:"storageScope"`
	DocumentType   string `json:"documentType,omitempty"`
	DocumentName   string `json:"documentName,omitempty"`
	DocumentStatus string `json:"documentStatus"`
	SortOrder      int    `json:"sortOrder"`
	FileUploaded   int    `json:"fileUploaded"`
	FilePath       string `json:"filePath,omitempty"`
	FileUploadedAt string `json:"fileUploadedAt,omitempty"`
	CheckResult    string `json:"checkResult,omitempty"`
	CheckNote      string `json:"checkNote,omitempty"`
	Files          []File `json:"files,omitempty"`
}
type UploadResponse struct {
	RequestID     string `json:"requestId,omitempty"`
	RequestNo     string `json:"requestNo,omitempty"`
	ProfileID     string `json:"profileRequestItemId,omitempty"`
	RequestFileID string `json:"requestItemFileId,omitempty"`
	ItemCode      string `json:"itemCode"`
	StorageScope  string `json:"storageScope"`
	DocumentType  string `json:"documentType"`
	Complete      bool   `json:"complete"`
	Files         []File `json:"files"`
}
type PreviewResponse struct {
	ItemCode     string `json:"documentMasterRequestItemCode"`
	DocumentType string `json:"documentType,omitempty"`
	Files        []File `json:"files"`
}
