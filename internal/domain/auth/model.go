package auth

type User struct {
	MobileUUID string
	Username   string
	Password   string
	Email      string
	Status     string
	FirstName  string
	LastName   string
}

type RegisterRequest struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	CompanyCode  string `json:"companyCode"`
	DateOfBirth  string `json:"dateOfBirth"`
	PositionCode string `json:"positionCode"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	MobileNumber string `json:"mobileNumber"`
}

type ResetPasswordRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordRequest struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	UID             string `json:"uid"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginResponse struct {
	Token             string `json:"token"`
	RefreshToken      string `json:"refToken"`
	Username          string `json:"username"`
	LastLoginDateTime string `json:"lastLoginDateTime"`
}

type RegisterResponse struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	MobileNumber string `json:"mobileNumber,omitempty"`
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
}
