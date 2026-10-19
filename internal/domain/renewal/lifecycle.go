package renewal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	payments "github.com/smartseaman/mobile-api/internal/domain/payment"
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
type PaymentRequest = payments.Input
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
	Step                     int    `json:"step"`
	DocumentMobileStatusCode string `json:"documentMobileStatusCode"`
	NameTh                   string `json:"nameTh"`
	NameEn                   string `json:"nameEn"`
	State                    string `json:"state"`
}
type StageMobileStatus struct {
	DocumentMobileStatusCode string `json:"documentMobileStatusCode"`
	NameTh                   string `json:"nameTh"`
	NameEn                   string `json:"nameEn"`
	Step                     int    `json:"step"`
}
type StageCurrentStatus struct {
	ID                 string             `json:"id"`
	DocumentStatusCode string             `json:"documentStatusCode"`
	NameTh             string             `json:"nameTh"`
	NameEn             string             `json:"nameEn"`
	CSSColor           string             `json:"cssColor"`
	Step               *int               `json:"step"`
	MobileStatus       *StageMobileStatus `json:"mobileStatus"`
}
type StageResponse struct {
	RequestNo     string             `json:"requestNo"`
	CurrentStatus StageCurrentStatus `json:"currentStatus"`
	Stages        []Stage            `json:"stages"`
}
type PaymentResponse = payments.Response

type requestRow struct {
	ID, No, UserID, Mobile, Email, Code, StatusID, StatusCode, StatusName, IdempotencyKey string
	Amount                                                                                float64
	IsResubmit                                                                            bool
	Submitted                                                                             time.Time
}

func (r *Repository) stageCurrentStatus(ctx context.Context, requestID string) (StageCurrentStatus, error) {
	var result StageCurrentStatus
	var mobileCode, mobileTh, mobileEn sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT s.id,s.document_status_code,COALESCE(s.name_th,''),COALESCE(s.name_en,''),COALESCE(s.css_color,''),COALESCE(s.document_mobile_status_code,''),COALESCE(s.document_mobile_status_name_th,''),COALESCE(s.document_mobile_status_name_en,'') FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.id=?`, requestID).
		Scan(&result.ID, &result.DocumentStatusCode, &result.NameTh, &result.NameEn, &result.CSSColor, &mobileCode, &mobileTh, &mobileEn)
	if err != nil {
		return StageCurrentStatus{}, err
	}
	if step := progress(mobileCode.String); step > 0 {
		result.Step = &step
		result.MobileStatus = &StageMobileStatus{DocumentMobileStatusCode: mobileCode.String, NameTh: mobileTh.String, NameEn: mobileEn.String, Step: step}
	}
	return result, nil
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
	return r.validateResponseWithDocument(ctx, row, code, idempotency, name, nameTh, nameEn)
}

func (r *Repository) validateResponseWithDocument(ctx context.Context, row requestRow, code, idempotency, name, nameTh, nameEn string) (ValidateResponse, error) {
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
	addressErr := addressRows.Err()
	addressRows.Close()
	if addressErr != nil {
		return ValidateResponse{}, addressErr
	}
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
		response.Items = append(response.Items, item)
	}
	itemsErr := items.Err()
	items.Close()
	if itemsErr != nil {
		return ValidateResponse{}, itemsErr
	}
	if len(response.Items) == 0 {
		return ValidateResponse{}, &BusinessError{Code: "MA00013", Detail: "documentRenewalRequestItems"}
	}
	// Release item rows before loading files, so even a one-connection pool works.
	files, err := r.validateFilesBatch(ctx, row, response.Items)
	if err != nil {
		return ValidateResponse{}, err
	}
	for index := range response.Items {
		item := &response.Items[index]
		item.Files = files[validateFileKey{scope: validateFileScope(item.StorageScope), code: item.ItemCode}]

		item.DocumentStatus = validateItemStatus(*item, item.Files)
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

type validateFileKey struct{ scope, code string }

// Preserve the existing fallback: non-REQUEST scopes read profile files.
func validateFileScope(scope string) string {
	if scope == "REQUEST" {
		return "REQUEST"
	}
	return "PROFILE"
}

// At most one file query per scope, independent of the number of items.
func (r *Repository) validateFilesBatch(ctx context.Context, row requestRow, items []ValidateItem) (map[validateFileKey][]ValidateFile, error) {
	result := make(map[validateFileKey][]ValidateFile, len(items))
	scopes := map[string][]string{}
	for _, item := range items {
		scope := validateFileScope(item.StorageScope)
		key := validateFileKey{scope: scope, code: item.ItemCode}
		if _, exists := result[key]; exists {
			continue
		}
		result[key] = []ValidateFile{}
		scopes[scope] = append(scopes[scope], item.ItemCode)
	}
	for _, scope := range []string{"PROFILE", "REQUEST"} {
		codes := scopes[scope]
		if len(codes) == 0 {
			continue
		}
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(codes)), ",")
		query := `SELECT document_master_request_item_code,id,document_type,slot_code,COALESCE(file_path,''),COALESCE(original_file_name,''),COALESCE(mime_type,''),COALESCE(file_size,0),COALESCE(DATE_FORMAT(file_uploaded_at,'%d/%m/%Y %H:%i'),''),COALESCE(check_result,''),COALESCE(check_note,''),COALESCE(is_updated,0) FROM m_document_profile_request_item WHERE mobile_user_uuid=? AND document_master_request_item_code IN (` + placeholders + `) AND file_uploaded=1 ORDER BY file_uploaded_at,id`
		args := []any{row.UserID}
		if scope == "REQUEST" {
			query = `SELECT i.document_master_request_item_code,f.id,f.document_type,f.slot_code,COALESCE(f.file_path,''),COALESCE(f.original_file_name,''),COALESCE(f.mime_type,''),COALESCE(f.file_size,0),COALESCE(DATE_FORMAT(f.file_uploaded_at,'%d/%m/%Y %H:%i'),''),COALESCE(f.check_result,''),COALESCE(f.check_note,''),COALESCE(f.is_updated,0) FROM m_document_request_item_files f INNER JOIN m_document_request_items i ON i.id=f.request_item_id INNER JOIN m_document_request r ON r.id=i.request_id WHERE i.request_id=? AND r.mobile_user_uuid=? AND i.document_master_request_item_code IN (` + placeholders + `) AND f.file_uploaded=1 ORDER BY f.file_uploaded_at,f.id`
			args = []any{row.ID, row.UserID}
		}
		for _, code := range codes {
			args = append(args, code)
		}
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var code string
			var file ValidateFile
			if err := rows.Scan(&code, &file.FileID, &file.DocumentType, &file.SlotCode, &file.FilePath, &file.OriginalName, &file.MimeType, &file.FileSize, &file.UploadedAt, &file.CheckResult, &file.CheckNote, &file.IsUpdated); err != nil {
				rows.Close()
				return nil, err
			}
			key := validateFileKey{scope: scope, code: code}
			result[key] = append(result[key], file)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
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
	var address ValidateAddress
	addressErr := r.db.QueryRowContext(ctx, `SELECT id,first_name,last_name,address_line,province,district,sub_district,postal_code FROM m_delivery_address WHERE mobile_user_uuid=? AND is_default=1 AND is_active='YES' ORDER BY updated_at DESC,id LIMIT 1`, userID).Scan(&address.ID, &address.FirstName, &address.LastName, &address.AddressLine, &address.Province, &address.District, &address.SubDistrict, &address.PostalCode)
	var addressID any
	if addressErr == nil {
		addressID = address.ID
	}
	if addressErr != nil && !errors.Is(addressErr, sql.ErrNoRows) {
		return ValidateResponse{}, addressErr
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
	row := requestRow{ID: id, No: requestNo, UserID: userID, Mobile: mobile, Email: email, Code: code, IdempotencyKey: idempotency}
	response, err := r.validateResponseWithDocument(ctx, row, code, idempotency, name, nameTh, nameEn)
	if err != nil {
		return ValidateResponse{}, err
	}
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
	row, err := r.ownedRequest(ctx, userID, "request_no", requestNo)
	if err != nil {
		return TransitionResponse{}, err
	}
	status, err := r.statusID(ctx, to)
	if err != nil {
		return TransitionResponse{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TransitionResponse{}, err
	}
	defer tx.Rollback()
	var current string
	if err = tx.QueryRowContext(ctx, "SELECT document_status_id FROM m_document_request WHERE id=? AND mobile_user_uuid=? AND is_active='YES' FOR UPDATE", row.ID, userID).Scan(&current); err != nil {
		return TransitionResponse{}, err
	}
	if current != row.StatusID {
		return TransitionResponse{}, &BusinessError{Code: "MA00005", Detail: "documentStatus"}
	}
	if action == "CANCEL" {
		var blocking int
		if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM mobile_payment_attempts WHERE request_id=? AND (status NOT IN ('EXPIRED','FAILED') OR review_required=1)", row.ID).Scan(&blocking); err != nil {
			return TransitionResponse{}, err
		}
		if blocking > 0 {
			return TransitionResponse{}, payments.Conflict("PAYMENT_IN_PROGRESS")
		}
	}
	activeClause := ""
	if action == "CANCEL" {
		activeClause = ",is_active='NO'"
	}
	if _, err = tx.ExecContext(ctx, "UPDATE m_document_request SET document_status_id=?,is_resubmit=CASE WHEN ?='RESUBMIT' THEN 1 ELSE is_resubmit END,updated_at=NOW()"+activeClause+" WHERE id=?", status, action, row.ID); err != nil {
		return TransitionResponse{}, err
	}
	note := to
	if action == "RESUBMIT" {
		note = "Corrected documents resubmitted"
	}
	if action == "CANCEL" {
		note = "Renewal request deleted by applicant"
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO m_document_transaction(id,request_id,action,from_status,to_status,note,actioned_by) VALUES(?,?,?,?,?,?,?)", uuid.NewString(), row.ID, action, row.StatusCode, to, note, userID); err != nil {
		return TransitionResponse{}, err
	}
	if err = tx.Commit(); err != nil {
		return TransitionResponse{}, err
	}
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
	var lockedID string
	if err = tx.QueryRowContext(ctx, "SELECT id FROM m_document_request WHERE id=? AND mobile_user_uuid=? FOR UPDATE", row.ID, userID).Scan(&lockedID); err != nil {
		return TransitionResponse{}, err
	}
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
	queries := []string{
		"DELETE FROM mobile_payment_request_state WHERE request_id=?",
		"DELETE FROM mobile_payment_attempts WHERE request_id=?",
		"DELETE f FROM m_document_request_item_files f INNER JOIN m_document_request_items i ON i.id=f.request_item_id WHERE i.request_id=?",
		"DELETE FROM m_document_request_items WHERE request_id=?",
		"DELETE FROM m_document_request_delivery_address WHERE request_id=?",
		"UPDATE m_payment_transaction SET parent_transaction_id=NULL WHERE request_id=?",
		"DELETE FROM m_document_transaction WHERE request_id=?",
		"DELETE FROM m_payment_transaction WHERE request_id=?",
		"DELETE FROM m_document_request WHERE id=? AND mobile_user_uuid=?",
	}
	for _, query := range queries {
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
