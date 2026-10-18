package document

import "context"

type CreateRequest struct {
	DocumentCode    string `json:"documentCode"`
	CertStartDate   string `json:"certStartDate"`
	CertEndDateType string `json:"certEndDateType"`
	CertEndDate     string `json:"certEndDate"`
	FileCert        string `json:"fileCert"`
	FileCertName    string `json:"fileCertName"`
}

type UpdateRequest struct {
	IsChangeFile    string `json:"isChangeFile"`
	DocumentCode    string `json:"documentCode"`
	CertStartDate   string `json:"certStartDate"`
	CertEndDateType string `json:"certEndDateType"`
	CertEndDate     string `json:"certEndDate"`
	FileCert        string `json:"fileCert"`
	FileCertName    string `json:"fileCertName"`
}

type Response struct {
	DocumentCode    string `json:"documentCode"`
	CertStartDate   string `json:"certStartDate"`
	CertEndDateType string `json:"certEndDateType"`
	CertEndDate     string `json:"certEndDate"`
	FileCertName    string `json:"fileCertName"`
}

type Item struct {
	DocumentID                 string `json:"documentId"`
	DocumentCode               string `json:"documentCode"`
	DocumentNameTh             string `json:"documentNameTh"`
	DocumentNameEn             string `json:"documentNameEn"`
	DocumentDescription        string `json:"documentDescription"`
	DocumentType               string `json:"documentType"`
	DocumentDefaultFlag        string `json:"documentDefaultFlag"`
	DocumentRenewalFlag        string `json:"documentRenewalFlag"`
	DocumentSeq                string `json:"documentSeq"`
	DocumentMobileFlag         string `json:"documentMobileFlag"`
	DocumentCourseCode         string `json:"documentCourseCode"`
	CertStartDate              string `json:"certStartDate"`
	CertEndDate                string `json:"certEndDate"`
	CertFileName               string `json:"certFileName"`
	DisYear                    string `json:"disYear"`
	DisMonth                   string `json:"disMonth"`
	DisDay                     string `json:"disDay"`
	CertExpiredFlag            string `json:"certExpiredFlag"`
	DocumentRenewalRequestFlag string `json:"documentRenewalRequestFlag"`
}

type PageResponse struct {
	ItemTotal int    `json:"itemTotal"`
	Last      bool   `json:"last"`
	Items     []Item `json:"items"`
}

type Storage interface {
	Put(context.Context, string, string, []byte) error
	Get(context.Context, string) ([]byte, string, error)
}
