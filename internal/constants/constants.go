package constants

// API response codes shared by HTTP handlers and domain services.
const (
	CodeSuccess       = "MA00000"
	CodeUnauthorized  = "MA00004"
	CodeInvalidInput  = "MA00005"
	CodeNotFound      = "MA00016"
	CodeTechnical     = "MA00013"
	CodeMissingHeader = "MA00011"
	CodeInvalidToken  = "MA00007"
)

// Document upload types and slots accepted by the mobile API.
const (
	DocumentTypeGeneral  = "GENERAL"
	DocumentTypeIDCard   = "ID_CARD"
	DocumentTypePassport = "PASSPORT"

	DocumentSlotMain  = "MAIN"
	DocumentSlotFront = "FRONT"
	DocumentSlotBack  = "BACK"
)

// Renewal statuses used by upload and lifecycle validation.
const (
	StatusPaymentPending        = "PAYMENT_PENDING"
	StatusApplicantCorrection   = "PENDING_APPLICANT_CORRECTION"
	RequestItemApprovalFix      = "FIX"
	DocumentStorageScopeProfile = "PROFILE"
	DocumentStorageScopeRequest = "REQUEST"
)
