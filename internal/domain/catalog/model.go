package catalog

import "context"

type Storage interface {
	Get(context.Context, string) ([]byte, string, error)
}
type Banner struct {
	BannerID       string `json:"bannerId"`
	BannerFileName string `json:"bannerFileName"`
}
type BannerResponse struct {
	Banners []Banner `json:"banners"`
}
type Voucher struct {
	VoucherID string `json:"voucherId"`
	Title     string `json:"voucherTitle"`
	Details   string `json:"voucherDetails"`
	StartDate string `json:"voucherStartDate"`
	EndDate   string `json:"voucherEndDate"`
}
type VoucherResponse struct {
	Vouchers []Voucher `json:"voucherModel"`
}
type School struct {
	CompanyCode      string `json:"companyCode"`
	CompanyNameEn    string `json:"companyNameEn"`
	CompanyNameTh    string `json:"companyNameTh"`
	CompanyColour    string `json:"companyColour"`
	CourseColour     string `json:"courseColour"`
	CourseID         string `json:"courseId"`
	CourseCode       string `json:"courseCode"`
	CourseNameEn     string `json:"courseNameEn"`
	CourseNameTh     string `json:"courseNameTh"`
	CourseType       string `json:"courseType"`
	CourseOnlineDate string `json:"courseOnlineDate"`
	CourseOnsiteDate string `json:"courseOnsiteDate"`
	CourseTotalDays  string `json:"courseTotalDays"`
	CoursePrice      string `json:"coursePrice"`
	CompanyLine      string `json:"companyLine"`
	CompanyFacebook  string `json:"companyFacebook"`
	CompanyPhone1    string `json:"companyPhone1"`
	CourseStartDate  string `json:"courseStartDate"`
	CourseEndDate    string `json:"courseEndDate"`
	DateForCheck     string `json:"dateForCheck"`
}
type SchoolResponse struct {
	Schools []School `json:"schoolTrainings"`
}
