package renewal

// Document request status codes shared with the legacy system and stored in
// m_document_status and m_document_transaction.from_status/to_status.
const (
	// StatusPaymentPending: รอชำระเงิน.
	StatusPaymentPending = "PAYMENT_PENDING"
	// StatusPendingDocumentReview: รอตรวจเอกสาร.
	StatusPendingDocumentReview = "PENDING_DOCUMENT_REVIEW"
	// StatusPendingApplicantCorrection: รอผู้ยื่นแก้ไข.
	StatusPendingApplicantCorrection = "PENDING_APPLICANT_CORRECTION"
	// StatusPendingMarineDepartmentResult: รอผลกรมเจ้าท่า.
	StatusPendingMarineDepartmentResult = "PENDING_MARINE_DEPARTMENT_RESULT"
	// StatusPendingDepartmentDocumentPickup: รอรับเอกสารจากกรม.
	StatusPendingDepartmentDocumentPickup = "PENDING_DEPARTMENT_DOCUMENT_PICKUP"
	// StatusDelivering: กำลังจัดส่ง.
	StatusDelivering = "DELIVERING"
	// StatusDelivered: จัดส่งสำเร็จ.
	StatusDelivered = "DELIVERED"
	// StatusCancelled: ยกเลิก.
	StatusCancelled = "CANCELLED"
)
