package renewal

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("renewal request not found")

type Repository struct {
	db    *sql.DB
	omise *omiseClient
}

func NewRepository(db *sql.DB, omise *omiseClient) *Repository {
	return &Repository{db: db, omise: omise}
}

func (r *Repository) Omise() *omiseClient { return r.omise }

type statusRow struct{ ID, Code, MobileCode, MobileNameTh, MobileNameEn, NameTh, NameEn, CSSColor string }

func (r *Repository) Statuses(ctx context.Context) ([]statusRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, document_status_code, document_mobile_status_code, document_mobile_status_name_th,
		document_mobile_status_name_en, name_th, name_en, css_color FROM m_document_status
		WHERE is_active='YES' AND is_mobile_visible='YES'
		ORDER BY CASE document_status_code WHEN 'PENDING_DOCUMENT_REVIEW' THEN 1 WHEN 'PENDING_APPLICANT_CORRECTION' THEN 2
		WHEN 'PENDING_MARINE_DEPARTMENT_RESULT' THEN 3 WHEN 'PENDING_DEPARTMENT_DOCUMENT_PICKUP' THEN 4
		WHEN 'DELIVERING' THEN 5 WHEN 'DELIVERED' THEN 6 WHEN 'CANCELLED' THEN 7 ELSE 99 END, name_th`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]statusRow, 0)
	for rows.Next() {
		var row statusRow
		if err := rows.Scan(&row.ID, &row.Code, &row.MobileCode, &row.MobileNameTh, &row.MobileNameEn, &row.NameTh, &row.NameEn, &row.CSSColor); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *Repository) Documents(ctx context.Context) ([]Document, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT DOCUMENT_CODE, DOCUMENT_NAME_TH, DOCUMENT_NAME_EN FROM m_documents WHERE DOCUMENT_STATUS='A' AND DOCUMENT_RENEWAL_FLAG IN ('Y','YES') ORDER BY DOCUMENT_SEQ")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Document, 0)
	for rows.Next() {
		var code, th, en sql.NullString
		if err := rows.Scan(&code, &th, &en); err != nil {
			return nil, err
		}
		result = append(result, Document{DocumentCode: code.String, DocumentName: th.String, DocumentNameTh: th.String, DocumentNameEn: en.String})
	}
	return result, rows.Err()
}

func (r *Repository) Price(ctx context.Context, code string) (Price, error) {
	var price Price
	var government, processing, shipping, shippingDiscount, serviceDiscount sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `SELECT p.document_code, p.government_fee, p.document_processing_fee, p.shipping_fee,
		p.shipping_discount, p.service_fee_discount, p.effective_from, p.effective_to
		FROM m_document_prices_setting p INNER JOIN m_documents d ON d.DOCUMENT_CODE=p.document_code
		WHERE p.document_code=? AND p.is_active='YES' AND p.effective_from<=CURRENT_DATE
		AND (p.effective_to IS NULL OR p.effective_to>=CURRENT_DATE) AND d.DOCUMENT_STATUS='A'
		ORDER BY p.effective_from DESC LIMIT 1`, code).Scan(&price.DocumentCode, &government, &processing, &shipping, &shippingDiscount, &serviceDiscount, &price.EffectiveFrom, &price.EffectiveTo)
	if err != nil {
		return Price{}, err
	}
	price.GovernmentFee, price.DocumentProcessingFee, price.ShippingFee = government.Float64, processing.Float64, shipping.Float64
	price.ShippingDiscount, price.ServiceFeeDiscount = shippingDiscount.Float64, serviceDiscount.Float64
	price.Total = price.GovernmentFee + price.DocumentProcessingFee + price.ShippingFee - price.ShippingDiscount - price.ServiceFeeDiscount
	return price, nil
}
