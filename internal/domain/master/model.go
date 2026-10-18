package master

type Company struct {
	Code string `json:"companyCode"`
	Name string `json:"companyName"`
}

type Position struct {
	Code string `json:"positionCode"`
	Name string `json:"positionName"`
}

type MasterResponse struct {
	Company  []Company  `json:"company"`
	Position []Position `json:"position"`
}

type Document struct {
	Code   string `json:"documentCode"`
	Name   string `json:"documentName"`
	NameTH string `json:"documentNameTh"`
}

type DocumentsResponse struct {
	Documents []Document `json:"documents"`
	COT       []Document `json:"cot"`
}

type Course struct {
	Code   string `json:"courseCode"`
	NameEN string `json:"courseNameEn"`
	NameTH string `json:"courseNameTh"`
}

type CoursesResponse struct {
	Courses []Course `json:"courses"`
}

type Address struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NameTH     string `json:"nameTh"`
	NameEN     string `json:"nameEn"`
	PostalCode string `json:"postalCode,omitempty"`
}
