package renewal

import "context"

// requestItems builds the shared live item/file representation for validation and timeline status.
func (r *Repository) requestItems(ctx context.Context, row requestRow, code string) ([]ValidateItem, error) {
	result := []ValidateItem{}
	items, err := r.db.QueryContext(ctx, `SELECT i.id,COALESCE(p.id,''),m.document_master_items_code,m.storage_scope,COALESCE(p.document_type,''),COALESCE(m.document_master_items_name,''),m.sort_order,CASE WHEN COALESCE(p.file_uploaded,0)=1 THEN 1 ELSE 0 END,CASE WHEN COALESCE(p.file_uploaded,0)=1 THEN 'COMPLETE' WHEN p.id IS NULL AND m.storage_scope='PROFILE' THEN 'MISSING' ELSE 'INCOMPLETE' END,COALESCE(p.file_path,''),COALESCE(DATE_FORMAT(p.file_uploaded_at,'%Y-%m-%dT%H:%i:%sZ'),''),COALESCE(p.check_result,''),COALESCE(p.check_note,'') FROM m_document_request_items i INNER JOIN m_document_master_request_item m ON m.document_master_items_code=i.document_master_request_item_code LEFT JOIN m_document_profile_request_item p ON p.id=(SELECT MIN(p2.id) FROM m_document_profile_request_item p2 WHERE p2.mobile_user_uuid=? AND p2.document_master_request_item_code=i.document_master_request_item_code) WHERE i.request_id=? ORDER BY m.sort_order,i.id`, row.UserID, row.ID)
	if err != nil {
		return nil, err
	}
	for items.Next() {
		var item ValidateItem
		if err := items.Scan(&item.ID, &item.ProfileID, &item.ItemCode, &item.StorageScope, &item.DocumentType, &item.DocumentName, &item.SortOrder, &item.FileUploaded, &item.DocumentStatus, &item.FilePath, &item.FileUploadedAt, &item.CheckResult, &item.CheckNote); err != nil {
			items.Close()
			return nil, err
		}
		item.DocumentCode = code
		result = append(result, item)
	}
	itemsErr := items.Err()
	items.Close()
	if itemsErr != nil {
		return nil, itemsErr
	}
	if len(result) == 0 {
		return nil, &BusinessError{Code: "MA00013", Detail: "documentRenewalRequestItems"}
	}
	// Release item rows before loading files, so even a one-connection pool works.
	files, err := r.validateFilesBatch(ctx, row, result)
	if err != nil {
		return nil, err
	}
	for index := range result {
		item := &result[index]
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
	return result, nil
}
