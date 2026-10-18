package renewal

import "time"

type Status struct {
	ID                 string        `json:"id"`
	DocumentStatusCode string        `json:"documentStatusCode"`
	NameTh             string        `json:"nameTh"`
	NameEn             string        `json:"nameEn"`
	CSSColor           string        `json:"cssColor"`
	ProgressStep       *int          `json:"progressStep"`
	MobileStatus       *MobileStatus `json:"mobileStatus"`
	Correction         bool          `json:"correction"`
	Terminal           bool          `json:"terminal"`
}
type MobileStatus struct {
	Code   string `json:"documentMobileStatusCode"`
	NameTh string `json:"nameTh"`
	NameEn string `json:"nameEn"`
	Step   int    `json:"step"`
}
type Document struct {
	DocumentCode   string `json:"documentCode"`
	DocumentName   string `json:"documentName"`
	DocumentNameTh string `json:"documentNameTh"`
	DocumentNameEn string `json:"-"`
}
type Price struct {
	DocumentCode          string     `json:"documentCode"`
	GovernmentFee         float64    `json:"governmentFee"`
	DocumentProcessingFee float64    `json:"documentProcessingFee"`
	ShippingFee           float64    `json:"shippingFee"`
	ShippingDiscount      float64    `json:"shippingDiscount"`
	ServiceFeeDiscount    float64    `json:"serviceFeeDiscount"`
	Total                 float64    `json:"total"`
	EffectiveFrom         *time.Time `json:"effectiveFrom"`
	EffectiveTo           *time.Time `json:"effectiveTo"`
}
