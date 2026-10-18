package profile

type Response struct {
	FirstName           string `json:"firstName"`
	LastName            string `json:"lastName"`
	DateOfBirth         string `json:"dateOfBirth"`
	Age                 string `json:"age"`
	Mobile              string `json:"mobile"`
	Email               string `json:"email"`
	CompanyCode         string `json:"companyCode"`
	CompanyDescription  string `json:"companyDescription"`
	PositionCode        string `json:"positionCode"`
	PositionDescription string `json:"positionDescription"`
	SmartSeamanID       string `json:"smartSeamanId"`
	ShortName           string `json:"shortName"`
}

type UpdateRequest struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	CompanyCode  string `json:"companyCode"`
	DateOfBirth  string `json:"dateOfBirth"`
	PositionCode string `json:"positionCode"`
	Email        string `json:"email"`
	MobileNumber string `json:"mobileNumber"`
	IsChangeFile string `json:"isChangeFile"`
	ImageProfile string `json:"imageProfile"`
}

type ActivateRequest struct {
	Email string `json:"email"`
}
type RegisterResponse struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	MobileNumber string `json:"mobileNumber,omitempty"`
}

type User struct {
	UUID, SmartID, Username, Email, FirstName, LastName       string
	Mobile, CompanyCode, PositionCode, ProfilePicture, Status string
	DateOfBirth                                               string
}
