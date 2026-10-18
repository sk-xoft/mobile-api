package renewal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	DocumentCode      string `json:"documentCode"`
	DeliveryAddressID string `json:"deliveryAddressId"`
}
type MobileRequest struct {
	MobileNumber string `json:"mobileNumber"`
}
type PaymentRequest struct {
	PaymentMethod  string `json:"paymentMethod"`
	IdempotencyKey string `json:"idempotencyKey"`
	ReturnURI      string `json:"returnUri"`
	PlatformType   string `json:"platformType"`
}
type CreateResponse struct {
	RequestID         string  `json:"requestId"`
	RequestNo         string  `json:"requestNo"`
	DocumentCode      string  `json:"documentCode"`
	Status            string  `json:"status"`
	Amount            float64 `json:"amount"`
	DeliveryAddressID string  `json:"deliveryAddressId"`
}
type ValidateResponse struct {
	RequestID      string            `json:"requestId,omitempty"`
	RequestNo      string            `json:"requestNo,omitempty"`
	DocumentCode   string            `json:"documentCode"`
	DocumentName   string            `json:"documentName,omitempty"`
	DocumentNameTh string            `json:"documentNameTh,omitempty"`
	DocumentNameEn string            `json:"documentNameEn,omitempty"`
	IdempotencyKey *string           `json:"idempotencyKey"`
	MobileNumber   string            `json:"mobileNumber,omitempty"`
	Email          string            `json:"email,omitempty"`
	CertEndDate    *string           `json:"certEndDate"`
	Address        []ValidateAddress `json:"address"`
	Items          []ValidateItem    `json:"items"`
}
type ValidateAddress struct {
	ID          string `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	AddressLine string `json:"addressLine"`
	Province    string `json:"province"`
	District    string `json:"district"`
	SubDistrict string `json:"subDistrict"`
	PostalCode  string `json:"postalCode"`
	Mobile      string `json:"mobileNumber"`
	Description string `json:"description,omitempty"`
	IsDefault   *bool  `json:"isDefault"`
}
type ValidateItem struct {
	ID             string         `json:"id"`
	ProfileID      string         `json:"profileRequestItemId,omitempty"`
	DocumentCode   string         `json:"documentCode"`
	ItemCode       string         `json:"documentMasterRequestItemCode"`
	StorageScope   string         `json:"storageScope"`
	DocumentType   string         `json:"documentType,omitempty"`
	DocumentName   string         `json:"documentName,omitempty"`
	DocumentStatus string         `json:"documentStatus"`
	SortOrder      int            `json:"sortOrder"`
	FileUploaded   int            `json:"fileUploaded"`
	FilePath       string         `json:"filePath,omitempty"`
	FileUploadedAt string         `json:"fileUploadedAt,omitempty"`
	CheckResult    string         `json:"checkResult,omitempty"`
	CheckNote      string         `json:"checkNote,omitempty"`
	Files          []ValidateFile `json:"files,omitempty"`
}
type ValidateFile struct {
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
type TransitionResponse struct {
	RequestNo  string `json:"requestNo"`
	FromStatus string `json:"fromStatus"`
	ToStatus   string `json:"toStatus"`
	Action     string `json:"action"`
}
type Summary struct {
	RequestID    string  `json:"requestId"`
	RequestNo    string  `json:"requestNo"`
	DocumentCode string  `json:"documentCode"`
	Status       string  `json:"status"`
	StatusName   string  `json:"statusName"`
	Amount       float64 `json:"amount"`
	SubmittedAt  string  `json:"submittedAt"`
}
type Page struct {
	ItemTotal int       `json:"itemTotal"`
	Last      bool      `json:"last"`
	Items     []Summary `json:"items"`
}
type Detail struct {
	RequestID      string  `json:"requestId"`
	RequestNo      string  `json:"requestNo"`
	MobileUserUUID string  `json:"mobileUserUuid"`
	DocumentCode   string  `json:"documentCode"`
	Status         string  `json:"status"`
	Amount         float64 `json:"amount"`
	SubmittedAt    string  `json:"submittedAt"`
	IsResubmit     bool    `json:"isResubmit"`
}
type TimelineItem struct {
	Action     string `json:"action"`
	FromStatus string `json:"fromStatus"`
	ToStatus   string `json:"toStatus"`
	Note       string `json:"note"`
	ActionedAt string `json:"actionedAt"`
}
type Timeline struct {
	RequestNo string         `json:"requestNo"`
	Items     []TimelineItem `json:"items"`
}
type Stage struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Step    int    `json:"step"`
	Current bool   `json:"current"`
}
type StageResponse struct {
	RequestNo     string  `json:"requestNo"`
	CurrentStatus string  `json:"currentStatus"`
	Stages        []Stage `json:"stages"`
}
type PaymentResponse struct {
	RequestID        string  `json:"requestId"`
	RequestNo        string  `json:"requestNo"`
	TransactionID    string  `json:"transactionId"`
	TransactionNo    string  `json:"transactionNo"`
	Status           string  `json:"status"`
	Channel          string  `json:"channel"`
	PaymentMethod    string  `json:"paymentMethod"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Provider         string  `json:"provider"`
	ProviderChargeID string  `json:"providerChargeId,omitempty"`
	AuthorizeURI     string  `json:"authorizeUri,omitempty"`
	ReturnURI        string  `json:"returnUri,omitempty"`
	QRCodeURI        string  `json:"qrCodeDownloadUri,omitempty"`
	ExpiresAt        string  `json:"expiresAt,omitempty"`
	FailureCode      string  `json:"failureCode,omitempty"`
	FailureMessage   string  `json:"failureMessage,omitempty"`
}

type requestRow struct {
	ID, No, UserID, Mobile, Email, Code, StatusID, StatusCode, StatusName, IdempotencyKey string
	Amount                                                                                float64
	IsResubmit                                                                            bool
	Submitted                                                                             time.Time
}

func (r *Repository) ownedRequest(ctx context.Context, userID, key, value string) (requestRow, error) {
	var row requestRow
	err := r.db.QueryRowContext(ctx, `SELECT r.id,r.request_no,r.mobile_user_uuid,r.mobile_number,r.email,r.document_code,r.document_status_id,s.document_status_code,s.name_en,COALESCE(r.idempotency_key,''),r.amount,r.is_resubmit,r.submitted_at FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.`+key+`=? AND r.mobile_user_uuid=? AND r.is_active='YES'`, value, userID).Scan(&row.ID, &row.No, &row.UserID, &row.Mobile, &row.Email, &row.Code, &row.StatusID, &row.StatusCode, &row.StatusName, &row.IdempotencyKey, &row.Amount, &row.IsResubmit, &row.Submitted)
	if errors.Is(err, sql.ErrNoRows) {
		return requestRow{}, ErrNotFound
	}
	return row, err
}
func (r *Repository) latestActive(ctx context.Context, userID, code string) (requestRow, error) {
	var row requestRow
	err := r.db.QueryRowContext(ctx, `SELECT r.id,r.request_no,r.mobile_user_uuid,r.mobile_number,r.email,r.document_code,r.document_status_id,s.document_status_code,s.name_en,COALESCE(r.idempotency_key,''),r.amount,r.is_resubmit,r.submitted_at FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.mobile_user_uuid=? AND r.document_code=? AND r.is_active='YES' AND s.document_status_code<>'DELIVERED' ORDER BY r.submitted_at DESC,r.id DESC LIMIT 1`, userID, code).Scan(&row.ID, &row.No, &row.UserID, &row.Mobile, &row.Email, &row.Code, &row.StatusID, &row.StatusCode, &row.StatusName, &row.IdempotencyKey, &row.Amount, &row.IsResubmit, &row.Submitted)
	if errors.Is(err, sql.ErrNoRows) {
		return requestRow{}, ErrNotFound
	}
	return row, err
}
func (r *Repository) statusID(ctx context.Context, code string) (string, error) {
	var id string
	e := r.db.QueryRowContext(ctx, "SELECT id FROM m_document_status WHERE document_status_code=? AND is_active='YES' LIMIT 1", code).Scan(&id)
	return id, e
}
func (r *Repository) create(ctx context.Context, userID string, req CreateRequest, price Price) (CreateResponse, error) {
	var mobile, email string
	if e := r.db.QueryRowContext(ctx, "SELECT MOBILE_NUMBER,EMAIL FROM m_mobile_users WHERE MOBILE_UUID=?", userID).Scan(&mobile, &email); e != nil {
		return CreateResponse{}, e
	}
	statusID, e := r.statusID(ctx, "PAYMENT_PENDING")
	if e != nil {
		return CreateResponse{}, e
	}
	id := uuid.NewString()
	period := time.Now().Format("0601")
	if _, e = r.db.ExecContext(ctx, "INSERT INTO m_document_request_sequence(period,last_number) VALUES(?,1) ON DUPLICATE KEY UPDATE last_number=last_number+1", period); e != nil {
		return CreateResponse{}, e
	}
	var number int
	if e = r.db.QueryRowContext(ctx, "SELECT last_number FROM m_document_request_sequence WHERE period=?", period).Scan(&number); e != nil {
		return CreateResponse{}, e
	}
	no := fmt.Sprintf("%s%05d", period, number)
	var address any
	if strings.TrimSpace(req.DeliveryAddressID) != "" {
		address = req.DeliveryAddressID
	}
	_, e = r.db.ExecContext(ctx, `INSERT INTO m_document_request(id,request_no,mobile_user_uuid,mobile_number,email,document_code,document_status_id,price_setting_id,delivery_address_id,is_active,amount) VALUES(?,?,?,?,?,?,?,NULL,?,'YES',?)`, id, no, userID, mobile, email, strings.ToUpper(strings.TrimSpace(req.DocumentCode)), statusID, address, price.Total)
	if e != nil {
		return CreateResponse{}, e
	}
	return CreateResponse{RequestID: id, RequestNo: no, DocumentCode: strings.ToUpper(strings.TrimSpace(req.DocumentCode)), Status: "PAYMENT_PENDING", Amount: price.Total, DeliveryAddressID: req.DeliveryAddressID}, nil
}

func (r *Repository) documentInfo(ctx context.Context, code string) (string, string, string, error) {
	var name, nameTh, nameEn string
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(DOCUMENT_NAME_TH,''),COALESCE(DOCUMENT_NAME_TH,''),COALESCE(DOCUMENT_NAME_EN,'') FROM m_documents WHERE DOCUMENT_CODE=? AND DOCUMENT_STATUS='A'", code).Scan(&name, &nameTh, &nameEn)
	return name, nameTh, nameEn, err
}

func (r *Repository) requiredItemCount(ctx context.Context, code string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_document_setting_requires WHERE document_code=? AND is_required=1 AND is_active='YES'", code).Scan(&count)
	return count, err
}

func (r *Repository) validateExisting(ctx context.Context, userID, key, value string) (requestRow, error) {
	return r.ownedRequest(ctx, userID, key, value)
}

func (r *Repository) validateResponse(ctx context.Context, row requestRow, code, idempotency string) (ValidateResponse, error) {
	name, nameTh, nameEn, err := r.documentInfo(ctx, code)
	if err != nil {
		return ValidateResponse{}, err
	}
	if idempotency == "" {
		idempotency = row.IdempotencyKey
	}
	response := ValidateResponse{RequestID: row.ID, RequestNo: row.No, DocumentCode: code, DocumentName: name, DocumentNameTh: nameTh, DocumentNameEn: nameEn, IdempotencyKey: stringPointer(idempotency), MobileNumber: row.Mobile, Email: row.Email, Address: []ValidateAddress{}, Items: []ValidateItem{}}
	var end sql.NullTime
	_ = r.db.QueryRowContext(ctx, "SELECT CERT_END_DATE FROM m_certificates WHERE CERT_MOBILE_UUID=? AND CERT_DOCUMENT_CODE=? ORDER BY CERT_END_DATE DESC LIMIT 1", row.UserID, code).Scan(&end)
	if end.Valid {
		formattedEndDate := end.Time.Format("2006-01-02")
		response.CertEndDate = &formattedEndDate
	}
	addressRows, err := r.db.QueryContext(ctx, `SELECT a.id,a.first_name,a.last_name,a.address_line,a.province,a.district,a.sub_district,a.postal_code,a.mobile_number,CONCAT_WS(' ',a.address_line,CONCAT('ตำบล',sd.name_in_thai),CONCAT('อำเภอ',d.name_in_thai),CONCAT('จังหวัด',p.name_in_thai),a.postal_code) FROM m_document_request_delivery_address a LEFT JOIN provinces p ON p.code=a.province LEFT JOIN districts d ON d.code=a.district LEFT JOIN subdistricts sd ON sd.code=a.sub_district WHERE a.request_id=? AND a.mobile_user_uuid=? ORDER BY a.created_at DESC,a.id DESC`, row.ID, row.UserID)
	if err != nil {
		return ValidateResponse{}, err
	}
	for addressRows.Next() {
		var a ValidateAddress
		if err := addressRows.Scan(&a.ID, &a.FirstName, &a.LastName, &a.AddressLine, &a.Province, &a.District, &a.SubDistrict, &a.PostalCode, &a.Mobile, &a.Description); err != nil {
			addressRows.Close()
			return ValidateResponse{}, err
		}
		response.Address = append(response.Address, a)
	}
	addressRows.Close()
	items, err := r.db.QueryContext(ctx, `SELECT i.id,COALESCE(p.id,''),m.document_master_items_code,m.storage_scope,COALESCE(p.document_type,''),COALESCE(m.document_master_items_name,''),m.sort_order,CASE WHEN COALESCE(p.file_uploaded,0)=1 THEN 1 ELSE 0 END,CASE WHEN COALESCE(p.file_uploaded,0)=1 THEN 'COMPLETE' WHEN p.id IS NULL AND m.storage_scope='PROFILE' THEN 'MISSING' ELSE 'INCOMPLETE' END,COALESCE(p.file_path,''),COALESCE(DATE_FORMAT(p.file_uploaded_at,'%Y-%m-%dT%H:%i:%sZ'),''),COALESCE(p.check_result,''),COALESCE(p.check_note,'') FROM m_document_request_items i INNER JOIN m_document_master_request_item m ON m.document_master_items_code=i.document_master_request_item_code LEFT JOIN m_document_profile_request_item p ON p.id=(SELECT MIN(p2.id) FROM m_document_profile_request_item p2 WHERE p2.mobile_user_uuid=? AND p2.document_master_request_item_code=i.document_master_request_item_code) WHERE i.request_id=? ORDER BY m.sort_order,i.id`, row.UserID, row.ID)
	if err != nil {
		return ValidateResponse{}, err
	}
	for items.Next() {
		var item ValidateItem
		if err := items.Scan(&item.ID, &item.ProfileID, &item.ItemCode, &item.StorageScope, &item.DocumentType, &item.DocumentName, &item.SortOrder, &item.FileUploaded, &item.DocumentStatus, &item.FilePath, &item.FileUploadedAt, &item.CheckResult, &item.CheckNote); err != nil {
			items.Close()
			return ValidateResponse{}, err
		}
		item.DocumentCode = code
		item.Files, err = r.validateFiles(ctx, row, item.ItemCode, item.StorageScope)
		if err != nil {
			items.Close()
			return ValidateResponse{}, err
		}
		item.DocumentStatus = validateItemStatus(item, item.Files)
		if item.StorageScope == "REQUEST" {
			item.FileUploaded = 0
			if len(item.Files) > 0 {
				item.FileUploaded = 1
				item.DocumentType = item.Files[0].DocumentType
				item.FilePath = item.Files[0].FilePath
				item.FileUploadedAt = item.Files[0].UploadedAt
				item.CheckResult = item.Files[0].CheckResult
				item.CheckNote = item.Files[0].CheckNote
			}
		}
		response.Items = append(response.Items, item)
	}
	items.Close()
	if len(response.Items) == 0 {
		return ValidateResponse{}, &BusinessError{Code: "MA00013", Detail: "documentRenewalRequestItems"}
	}
	return response, nil
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func validateItemStatus(item ValidateItem, files []ValidateFile) string {
	if len(files) == 0 {
		return "MISSING"
	}
	for _, file := range files {
		if strings.EqualFold(file.CheckResult, "fix") {
			return "NEED_FIX"
		}
	}
	if item.ItemCode == "MRI001" {
		front, back, main, passport := false, false, false, false
		for _, file := range files {
			if file.DocumentType == "ID_CARD" {
				front = front || file.SlotCode == "FRONT"
				back = back || file.SlotCode == "BACK"
				main = main || file.SlotCode == "MAIN"
			}
			passport = passport || file.DocumentType == "PASSPORT" && file.SlotCode == "MAIN"
		}
		if main || passport || front && back {
			return "COMPLETE"
		}
		return "INCOMPLETE"
	}
	for _, file := range files {
		if file.DocumentType == "GENERAL" && file.SlotCode == "MAIN" {
			return "COMPLETE"
		}
	}
	return "INCOMPLETE"
}

func (r *Repository) validateFiles(ctx context.Context, row requestRow, itemCode, storageScope string) ([]ValidateFile, error) {
	query := `SELECT id,document_type,slot_code,COALESCE(file_path,''),COALESCE(original_file_name,''),COALESCE(mime_type,''),COALESCE(file_size,0),COALESCE(DATE_FORMAT(file_uploaded_at,'%d/%m/%Y %H:%i'),''),COALESCE(check_result,''),COALESCE(check_note,''),COALESCE(is_updated,0) FROM m_document_profile_request_item WHERE mobile_user_uuid=? AND document_master_request_item_code=? AND file_uploaded=1 ORDER BY file_uploaded_at,id`
	args := []any{row.UserID, itemCode}
	if storageScope == "REQUEST" {
		query = `SELECT f.id,f.document_type,f.slot_code,COALESCE(f.file_path,''),COALESCE(f.original_file_name,''),COALESCE(f.mime_type,''),COALESCE(f.file_size,0),COALESCE(DATE_FORMAT(f.file_uploaded_at,'%d/%m/%Y %H:%i'),''),COALESCE(f.check_result,''),COALESCE(f.check_note,''),COALESCE(f.is_updated,0) FROM m_document_request_item_files f INNER JOIN m_document_request_items i ON i.id=f.request_item_id WHERE i.request_id=? AND i.document_master_request_item_code=? AND f.file_uploaded=1 ORDER BY f.file_uploaded_at,f.id`
		args = []any{row.ID, itemCode}
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := make([]ValidateFile, 0)
	for rows.Next() {
		var file ValidateFile
		if err := rows.Scan(&file.FileID, &file.DocumentType, &file.SlotCode, &file.FilePath, &file.OriginalName, &file.MimeType, &file.FileSize, &file.UploadedAt, &file.CheckResult, &file.CheckNote, &file.IsUpdated); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func (r *Repository) validateCreate(ctx context.Context, userID, code, idempotency string, price Price) (ValidateResponse, error) {
	var mobile, email string
	if err := r.db.QueryRowContext(ctx, "SELECT MOBILE_NUMBER,EMAIL FROM m_mobile_users WHERE MOBILE_UUID=?", userID).Scan(&mobile, &email); err != nil {
		return ValidateResponse{}, err
	}
	statusID, err := r.statusID(ctx, "PAYMENT_PENDING")
	if err != nil {
		return ValidateResponse{}, err
	}
	name, nameTh, nameEn, err := r.documentInfo(ctx, code)
	if err != nil {
		return ValidateResponse{}, err
	}
	itemCount, err := r.requiredItemCount(ctx, code)
	if err != nil || itemCount == 0 {
		if err != nil {
			return ValidateResponse{}, err
		}
		return ValidateResponse{}, &BusinessError{Code: "MA00013", Detail: "documentSetting"}
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ValidateResponse{}, err
	}
	defer tx.Rollback()
	id, period := uuid.NewString(), time.Now().Format("0601")
	if _, err = tx.ExecContext(ctx, "INSERT INTO m_document_request_sequence(period,last_number) VALUES(?,1) ON DUPLICATE KEY UPDATE last_number=last_number+1", period); err != nil {
		return ValidateResponse{}, err
	}
	var number int
	if err = tx.QueryRowContext(ctx, "SELECT last_number FROM m_document_request_sequence WHERE period=? FOR UPDATE", period).Scan(&number); err != nil {
		return ValidateResponse{}, err
	}
	requestNo := fmt.Sprintf("%s%05d", period, number)
	var address ValidateAddress
	addressErr := r.db.QueryRowContext(ctx, `SELECT id,first_name,last_name,address_line,province,district,sub_district,postal_code FROM m_delivery_address WHERE mobile_user_uuid=? AND is_default=1 AND is_active='YES' ORDER BY updated_at DESC,id LIMIT 1`, userID).Scan(&address.ID, &address.FirstName, &address.LastName, &address.AddressLine, &address.Province, &address.District, &address.SubDistrict, &address.PostalCode)
	var addressID any
	if addressErr == nil {
		addressID = address.ID
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO m_document_request(id,request_no,mobile_user_uuid,mobile_number,email,document_code,document_status_id,price_setting_id,delivery_address_id,is_active,amount,idempotency_key) VALUES(?,?,?,?,?,?,?,(SELECT id FROM m_document_prices_setting WHERE document_code=? AND is_active='YES' AND effective_from<=CURRENT_DATE AND (effective_to IS NULL OR effective_to>=CURRENT_DATE) ORDER BY effective_from DESC LIMIT 1),?,'YES',?,?)`, id, requestNo, userID, mobile, email, code, statusID, code, addressID, price.Total, nullString(idempotency))
	if err != nil {
		return ValidateResponse{}, err
	}
	if addressErr == nil {
		snapshotID := uuid.NewString()
		if _, err = tx.ExecContext(ctx, `INSERT INTO m_document_request_delivery_address(id,request_id,source_delivery_address_id,mobile_user_uuid,first_name,last_name,address_line,province,district,sub_district,postal_code,mobile_number) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, snapshotID, id, address.ID, userID, address.FirstName, address.LastName, address.AddressLine, address.Province, address.District, address.SubDistrict, address.PostalCode, mobile); err != nil {
			return ValidateResponse{}, err
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO m_document_request_items(id,request_id,request_no,document_master_request_item_code,approve_status) SELECT UUID(),?,?,document_master_request_item_code,'PENDING' FROM m_document_setting_requires WHERE document_code=? AND is_required=1 AND is_active='YES'`, id, requestNo, code); err != nil {
		return ValidateResponse{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO m_document_transaction(id,request_id,action,from_status,to_status,note,actioned_by) VALUES(?,?,?,?,?,?,?)`, uuid.NewString(), id, "CREATE", nil, "PAYMENT_PENDING", "Unpaid draft created", userID); err != nil {
		return ValidateResponse{}, err
	}
	if err = tx.Commit(); err != nil {
		return ValidateResponse{}, err
	}
	row, err := r.ownedRequest(ctx, userID, "id", id)
	if err != nil {
		return ValidateResponse{}, err
	}
	response, err := r.validateResponse(ctx, row, code, idempotency)
	if err != nil {
		return ValidateResponse{}, err
	}
	response.DocumentName, response.DocumentNameTh, response.DocumentNameEn = name, nameTh, nameEn
	return response, nil
}

func nullString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
func (r *Repository) list(ctx context.Context, userID string, offset int) ([]Summary, int, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT r.id,r.request_no,r.document_code,s.document_status_code,s.name_en,r.amount,r.submitted_at FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.mobile_user_uuid=? AND r.is_active='YES' ORDER BY r.submitted_at DESC LIMIT 10 OFFSET ?`, userID, offset)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()
	items := make([]Summary, 0)
	for rows.Next() {
		var i Summary
		var date time.Time
		if e := rows.Scan(&i.RequestID, &i.RequestNo, &i.DocumentCode, &i.Status, &i.StatusName, &i.Amount, &date); e != nil {
			return nil, 0, e
		}
		i.SubmittedAt = date.Format(time.RFC3339)
		items = append(items, i)
	}
	var total int
	e = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_document_request WHERE mobile_user_uuid=? AND is_active='YES'", userID).Scan(&total)
	return items, total, e
}
func (r *Repository) transition(ctx context.Context, userID, requestNo, to, action string) (TransitionResponse, error) {
	row, e := r.ownedRequest(ctx, userID, "request_no", requestNo)
	if e != nil {
		return TransitionResponse{}, e
	}
	if (action == "RESUBMIT" && row.StatusCode != "PENDING_APPLICANT_CORRECTION") ||
		(action == "CANCEL" && row.StatusCode != "PAYMENT_PENDING") {
		return TransitionResponse{}, &BusinessError{Code: "MA00005", Detail: "documentStatus"}
	}
	status, e := r.statusID(ctx, to)
	if e != nil {
		return TransitionResponse{}, e
	}
	activeClause := ""
	if action == "CANCEL" {
		activeClause = ",is_active='NO'"
	}
	if _, e = r.db.ExecContext(ctx, "UPDATE m_document_request SET document_status_id=?,is_resubmit=CASE WHEN ?='RESUBMIT' THEN 1 ELSE is_resubmit END,updated_at=NOW()"+activeClause+" WHERE id=?", status, action, row.ID); e != nil {
		return TransitionResponse{}, e
	}
	note := to
	if action == "RESUBMIT" {
		note = "Corrected documents resubmitted"
	}
	if action == "CANCEL" {
		note = "Renewal request deleted by applicant"
	}
	_, _ = r.db.ExecContext(ctx, "INSERT INTO m_document_transaction(id,request_id,action,from_status,to_status,note,actioned_by) VALUES(?,?,?,?,?,?,?)", uuid.NewString(), row.ID, action, row.StatusCode, to, note, userID)
	return TransitionResponse{RequestNo: row.No, FromStatus: row.StatusCode, ToStatus: to, Action: action}, nil
}

func (r *Repository) countIncompleteRequestItems(ctx context.Context, requestID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)
		FROM m_document_request_items i
		INNER JOIN m_document_master_request_item m ON m.document_master_items_code=i.document_master_request_item_code
		WHERE i.request_id=? AND m.storage_scope='REQUEST' AND NOT EXISTS (
			SELECT 1 FROM m_document_request_item_files f
			WHERE f.request_item_id=i.id AND f.file_uploaded=1
			AND (f.check_result IS NULL OR f.check_result<>'fix')
		)`, requestID).Scan(&count)
	return count, err
}

func (r *Repository) countFixItems(ctx context.Context, requestID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_document_request_items WHERE request_id=? AND approve_status='FIX'", requestID).Scan(&count)
	return count, err
}

func (r *Repository) countIncompleteFixItems(ctx context.Context, requestID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)
		FROM m_document_request_items i
		WHERE i.request_id=? AND i.approve_status='FIX' AND NOT EXISTS (
			SELECT 1 FROM m_document_request_item_files f
			WHERE f.request_item_id=i.id AND f.file_uploaded=1 AND f.is_updated=1
			AND (f.check_result IS NULL OR f.check_result<>'fix')
		)`, requestID).Scan(&count)
	return count, err
}

func (r *Repository) resetFixItems(ctx context.Context, requestID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE m_document_request_items SET approve_status='PENDING',updated_at=NOW() WHERE request_id=? AND approve_status='FIX'", requestID)
	return err
}
func (r *Repository) updateMobile(ctx context.Context, userID, no, mobile string) (TransitionResponse, error) {
	row, e := r.ownedRequest(ctx, userID, "request_no", no)
	if e != nil {
		return TransitionResponse{}, e
	}
	_, e = r.db.ExecContext(ctx, "UPDATE m_document_request SET mobile_number=?,updated_at=NOW() WHERE id=?", mobile, row.ID)
	return TransitionResponse{RequestNo: no, FromStatus: row.Mobile, ToStatus: mobile, Action: "UPDATE_MOBILE"}, e
}
func (r *Repository) detail(ctx context.Context, userID, no string) (Detail, error) {
	row, e := r.ownedRequest(ctx, userID, "request_no", no)
	if e != nil {
		return Detail{}, e
	}
	return Detail{RequestID: row.ID, RequestNo: row.No, MobileUserUUID: row.UserID, DocumentCode: row.Code, Status: row.StatusCode, Amount: row.Amount, SubmittedAt: row.Submitted.Format(time.RFC3339), IsResubmit: row.IsResubmit}, nil
}
func (r *Repository) timeline(ctx context.Context, userID, no string) (Timeline, error) {
	row, e := r.ownedRequest(ctx, userID, "request_no", no)
	if e != nil {
		return Timeline{}, e
	}
	rows, e := r.db.QueryContext(ctx, "SELECT action,from_status,to_status,note,actioned_at FROM m_document_transaction WHERE request_id=? ORDER BY actioned_at", row.ID)
	if e != nil {
		return Timeline{}, e
	}
	defer rows.Close()
	result := Timeline{RequestNo: no, Items: []TimelineItem{}}
	for rows.Next() {
		var i TimelineItem
		var at time.Time
		if e := rows.Scan(&i.Action, &i.FromStatus, &i.ToStatus, &i.Note, &at); e != nil {
			return Timeline{}, e
		}
		i.ActionedAt = at.Format(time.RFC3339)
		result.Items = append(result.Items, i)
	}
	return result, rows.Err()
}
func (r *Repository) payment(ctx context.Context, userID, requestID string, input PaymentRequest) (PaymentResponse, error) {
	row, e := r.ownedRequest(ctx, userID, "id", requestID)
	if e != nil {
		return PaymentResponse{}, e
	}
	var existing PaymentResponse
	var existingExpiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx, "SELECT id,transaction_no,status,channel,payment_method,amount,currency,provider,expires_at FROM m_payment_transaction WHERE request_id=? AND idempotency_key=?", requestID, input.IdempotencyKey).Scan(&existing.TransactionID, &existing.TransactionNo, &existing.Status, &existing.Channel, &existing.PaymentMethod, &existing.Amount, &existing.Currency, &existing.Provider, &existingExpiresAt)
	if err == nil {
		existing.RequestID = row.ID
		existing.RequestNo = row.No
		existing.ExpiresAt = formatNullableTime(existingExpiresAt)
		return existing, nil
	}
	if r.omise == nil || !r.omise.enabled() {
		return PaymentResponse{}, fmt.Errorf("OMISE_SECRET_KEY is not configured")
	}
	charge, e := r.omise.CreateCharge(row.Amount, row.No, input)
	if e != nil {
		return PaymentResponse{}, e
	}
	id := uuid.NewString()
	no := "PAY" + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))[:20]
	channel := "PROMPTPAY"
	if strings.HasPrefix(input.PaymentMethod, "mobile_banking_") {
		channel = "MOBILE_BANKING"
	}
	status := omisePaymentStatus(charge.Status)
	_, e = r.db.ExecContext(ctx, `INSERT INTO m_payment_transaction(id,request_id,transaction_no,transaction_type,channel,payment_method,amount,currency,status,provider,provider_charge_id,provider_source_id,provider_transaction_id,provider_status,provider_response,idempotency_key,description,return_uri,authorize_uri,failure_code,failure_message,is_livemode,expires_at) VALUES(?,?,?,'CHARGE',?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, id, requestID, no, channel, input.PaymentMethod, row.Amount, "THB", status, "OMISE", charge.ID, charge.SourceID, charge.TransactionID, charge.Status, charge.Raw, input.IdempotencyKey, "Document renewal "+row.No, charge.ReturnURI, charge.AuthorizeURI, charge.FailureCode, charge.FailureMessage, charge.Livemode, nullableTime(charge.ExpiresAt))
	return PaymentResponse{RequestID: row.ID, RequestNo: row.No, TransactionID: id, TransactionNo: no, Status: status, Channel: channel, PaymentMethod: input.PaymentMethod, Amount: row.Amount, Currency: "THB", Provider: "OMISE", ProviderChargeID: charge.ID, AuthorizeURI: charge.AuthorizeURI, ReturnURI: charge.ReturnURI, QRCodeURI: charge.QRCodeURI, ExpiresAt: formatTime(charge.ExpiresAt), FailureCode: charge.FailureCode, FailureMessage: charge.FailureMessage}, e
}

func omisePaymentStatus(status string) string {
	switch status {
	case "successful":
		return "SUCCESS"
	case "failed":
		return "FAILED"
	case "expired":
		return "EXPIRED"
	case "pending":
		return "PENDING"
	default:
		return "PROCESSING"
	}
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC()
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func formatNullableTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return formatTime(value.Time)
}

func (r *Repository) paymentStatus(ctx context.Context, userID, requestID, transactionID string) (PaymentResponse, error) {
	row, e := r.ownedRequest(ctx, userID, "id", requestID)
	if e != nil {
		return PaymentResponse{}, e
	}
	var p PaymentResponse
	var expiresAt sql.NullTime
	e = r.db.QueryRowContext(ctx, "SELECT id,transaction_no,status,channel,payment_method,amount,currency,provider,expires_at FROM m_payment_transaction WHERE id=? AND request_id=?", transactionID, requestID).Scan(&p.TransactionID, &p.TransactionNo, &p.Status, &p.Channel, &p.PaymentMethod, &p.Amount, &p.Currency, &p.Provider, &expiresAt)
	p.RequestID = row.ID
	p.RequestNo = row.No
	p.ExpiresAt = formatNullableTime(expiresAt)
	return p, e
}

func (r *Repository) hardDelete(ctx context.Context, userID, requestNo string) (TransitionResponse, error) {
	row, err := r.ownedRequest(ctx, userID, "request_no", requestNo)
	if err != nil {
		return TransitionResponse{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TransitionResponse{}, err
	}
	defer tx.Rollback()
	for _, table := range []string{"m_delivery", "m_dept_submission"} {
		var exists int
		if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?", table).Scan(&exists); err != nil {
			return TransitionResponse{}, fmt.Errorf("hard delete check table %s: %w", table, err)
		}
		if exists == 1 {
			if _, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE request_id=?", row.ID); err != nil {
				return TransitionResponse{}, fmt.Errorf("hard delete table %s: %w", table, err)
			}
		}
	}
	for _, query := range []string{
		"DELETE f FROM m_document_request_item_files f INNER JOIN m_document_request_items i ON i.id=f.request_item_id WHERE i.request_id=?",
		"DELETE FROM m_document_request_items WHERE request_id=?",
		"DELETE FROM m_document_request_delivery_address WHERE request_id=?",
		"UPDATE m_payment_transaction SET parent_transaction_id=NULL WHERE request_id=?",
		"DELETE FROM m_document_transaction WHERE request_id=?",
		"DELETE FROM m_payment_transaction WHERE request_id=?",
		"DELETE FROM m_document_request WHERE id=? AND mobile_user_uuid=?",
	} {
		var result sql.Result
		if query == "DELETE FROM m_document_request WHERE id=? AND mobile_user_uuid=?" {
			result, err = tx.ExecContext(ctx, query, row.ID, userID)
		} else {
			result, err = tx.ExecContext(ctx, query, row.ID)
		}
		if err != nil {
			return TransitionResponse{}, fmt.Errorf("hard delete query failed: %w", err)
		}
		if query == "DELETE FROM m_document_request WHERE id=? AND mobile_user_uuid=?" {
			n, _ := result.RowsAffected()
			if n != 1 {
				return TransitionResponse{}, ErrNotFound
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return TransitionResponse{}, err
	}
	return TransitionResponse{RequestNo: row.No, FromStatus: row.StatusCode, ToStatus: "DELETED", Action: "HARD_DELETE"}, nil
}

func (r *Repository) updateWebhook(ctx context.Context, charge omiseCharge) error {
	status := omisePaymentStatus(charge.Status)
	_, err := r.db.ExecContext(ctx, "UPDATE m_payment_transaction SET status=?, provider_source_id=?, provider_transaction_id=?, provider_status=?, provider_response=?, return_uri=?, authorize_uri=?, failure_code=?, failure_message=?, is_livemode=?, expires_at=COALESCE(?, expires_at), paid_at=CASE WHEN ?='SUCCESS' THEN NOW() ELSE paid_at END, failed_at=CASE WHEN ?='FAILED' THEN NOW() ELSE failed_at END, updated_at=NOW() WHERE provider='OMISE' AND provider_charge_id=?", status, charge.SourceID, charge.TransactionID, charge.Status, charge.Raw, charge.ReturnURI, charge.AuthorizeURI, charge.FailureCode, charge.FailureMessage, charge.Livemode, nullableTime(charge.ExpiresAt), status, status, charge.ID)
	if err != nil {
		return err
	}
	if status != "SUCCESS" {
		return nil
	}
	result, err := r.db.ExecContext(ctx, `UPDATE m_document_request r INNER JOIN m_payment_transaction p ON p.request_id=r.id SET r.document_status_id=(SELECT id FROM m_document_status WHERE document_status_code='PENDING_DOCUMENT_REVIEW' AND is_active='YES' LIMIT 1), r.updated_at=NOW() WHERE p.provider='OMISE' AND p.provider_charge_id=? AND r.document_status_id=(SELECT id FROM m_document_status WHERE document_status_code='PAYMENT_PENDING' AND is_active='YES' LIMIT 1)`, charge.ID)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil || updated != 1 {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO m_document_transaction(id,request_id,action,from_status,to_status,note,actioned_by)
		SELECT UUID(),p.request_id,'PAYMENT_SUCCESS','PAYMENT_PENDING','PENDING_DOCUMENT_REVIEW','Payment completed',r.mobile_user_uuid
		FROM m_document_request r INNER JOIN m_payment_transaction p ON p.request_id=r.id
		WHERE p.provider='OMISE' AND p.provider_charge_id=?`, charge.ID)
	return err
}
