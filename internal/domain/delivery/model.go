package delivery

type Request struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	AddressLine string `json:"addressLine"`
	Province    string `json:"province"`
	District    string `json:"district"`
	SubDistrict string `json:"subDistrict"`
	PostalCode  string `json:"postalCode"`
	IsDefault   *bool  `json:"isDefault"`
}

type Response struct {
	ID          string `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	AddressLine string `json:"addressLine"`
	Province    string `json:"province"`
	District    string `json:"district"`
	SubDistrict string `json:"subDistrict"`
	PostalCode  string `json:"postalCode"`
	Mobile      string `json:"mobileNumber"`
	Description string `json:"description"`
	IsDefault   bool   `json:"isDefault"`
}

type Address struct {
	ID, UserID, FirstName, LastName, AddressLine string
	Province, District, SubDistrict, PostalCode  string
	Mobile, Description                          string
	IsDefault                                    bool
}
