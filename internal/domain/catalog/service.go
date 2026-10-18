package catalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Service struct {
	repository                             repositoryPort
	storage                                Storage
	bannerPath, voucherPath, voucherQRPath string
}

type repositoryPort interface {
	Banners(context.Context) ([]Banner, error)
	Banner(context.Context, string) (Banner, error)
	UserSmartID(context.Context, string) (string, error)
	Vouchers(context.Context, string) ([]Voucher, error)
	Voucher(context.Context, string) (Voucher, string, string, error)
	Schools(context.Context, string, bool, string) ([]School, error)
}

func NewService(r repositoryPort, s Storage, bannerPath, voucherPath, voucherQRPath string) *Service {
	return &Service{repository: r, storage: s, bannerPath: strings.Trim(bannerPath, "/"), voucherPath: strings.Trim(voucherPath, "/"), voucherQRPath: strings.Trim(voucherQRPath, "/")}
}
func (s *Service) Banners(ctx context.Context) (BannerResponse, error) {
	v, e := s.repository.Banners(ctx)
	return BannerResponse{Banners: v}, e
}
func (s *Service) BannerImage(ctx context.Context, id string) ([]byte, error) {
	v, e := s.repository.Banner(ctx, id)
	if e != nil {
		return nil, e
	}
	return s.object(ctx, s.bannerPath+"/"+v.BannerFileName)
}
func (s *Service) Vouchers(ctx context.Context, userID string) (VoucherResponse, error) {
	smart, e := s.repository.UserSmartID(ctx, userID)
	if e != nil {
		return VoucherResponse{}, e
	}
	v, e := s.repository.Vouchers(ctx, smart)
	return VoucherResponse{Vouchers: v}, e
}
func (s *Service) Voucher(ctx context.Context, id string) (Voucher, string, string, error) {
	return s.repository.Voucher(ctx, id)
}
func (s *Service) VoucherImage(ctx context.Context, id string) ([]byte, error) {
	_, file, _, e := s.repository.Voucher(ctx, id)
	if e != nil {
		return nil, e
	}
	return s.object(ctx, s.voucherPath+"/"+file)
}
func (s *Service) VoucherQR(ctx context.Context, id string) ([]byte, error) {
	_, _, file, e := s.repository.Voucher(ctx, id)
	if e != nil {
		return nil, e
	}
	return s.object(ctx, s.voucherQRPath+"/"+file)
}
func (s *Service) Schools(ctx context.Context, course string) (SchoolResponse, error) {
	rows, e := s.repository.Schools(ctx, course, false, "")
	if e != nil {
		return SchoolResponse{}, e
	}
	uniq := map[string]School{}
	for _, row := range rows {
		key := row.CompanyCode + row.CourseCode
		if _, ok := uniq[key]; !ok {
			uniq[key] = row
		}
	}
	result := make([]School, 0, len(uniq))
	for _, row := range uniq {
		row.CourseStartDate, row.CourseEndDate = courseRange(row)
		result = append(result, row)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DateForCheck < result[j].DateForCheck })
	return SchoolResponse{Schools: result}, nil
}
func (s *Service) SchoolDetail(ctx context.Context, company, course string) (SchoolResponse, error) {
	v, e := s.repository.Schools(ctx, course, true, company)
	return SchoolResponse{Schools: v}, e
}
func (s *Service) object(ctx context.Context, key string) ([]byte, error) {
	content, _, e := s.storage.Get(ctx, key)
	if e != nil {
		return nil, e
	}
	return base64.StdEncoding.DecodeString(string(content))
}
func courseRange(row School) (string, string) {
	dates := []time.Time{}
	for _, value := range append(strings.Split(row.CourseOnlineDate, ","), strings.Split(row.CourseOnsiteDate, ",")...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if date, e := time.Parse("02/01/2006", value); e == nil {
			dates = append(dates, date)
		}
	}
	if len(dates) == 0 {
		return "", ""
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates[0].Format("02/01/2006"), dates[len(dates)-1].Format("02/01/2006")
}

type BusinessError struct{ Code, Detail string }

func (e *BusinessError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Detail) }
