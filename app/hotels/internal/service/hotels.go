package service

import (
	"context"

	v1 "yinni-travel-backend/api/hotels/v1"
	"yinni-travel-backend/app/hotels/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HotelsService struct {
	v1.UnimplementedHotelServer

	uc  *biz.HotelsUsecase
	log *log.Helper
}

func NewHotelsService(uc *biz.HotelsUsecase) *HotelsService {
	return &HotelsService{uc: uc}
}

func (s *HotelsService) ListHotels(ctx context.Context, req *v1.ListHotelsRequest) (*v1.ListHotelsReply, error) {
	if s.log != nil {
		s.log.WithContext(ctx).Infof("ListHotels called: page=%d, pageSize=%d", req.Page, req.PageSize)
	}

	params := &biz.ListHotelsParams{
		Page:        req.Page,
		PageSize:    req.PageSize,
		Category:    req.Category,
		Brand:       req.Brand,
		SubCategory: req.SubCategory,
		MinPrice:    req.MinPrice,
		MaxPrice:    req.MaxPrice,
		MinRating:   req.MinRating,
		InStock:     req.InStock,
		Featured:    req.FeaturedOnly,
		Seller:      req.Seller,
		SortBy:      req.SortBy,
		SortOrder:   req.SortOrder,
		SearchQuery: req.SearchQuery,
	}

	hotels, total, err := s.uc.GetHotelsList(ctx, params)
	if err != nil {
		if s.log != nil {
			s.log.WithContext(ctx).Errorf("ListHotels failed: %v", err)
		}
		return nil, err
	}

	return &v1.ListHotelsReply{
		Hotels:   s.convertToHotelsList(hotels),
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *HotelsService) SeedHotelsDatabase(ctx context.Context, req *v1.SeedHotelsDatabaseRequest) (*v1.SeedHotelsDatabaseReply, error) {
	result, err := s.uc.SeedHotelsDatabase(ctx, &biz.SeedHotelsParams{
		ClearExisting: req.GetClearExisting(),
		MaxHotels:     req.GetMaxHotels(),
		DatasetPath:   req.GetDatasetPath(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.SeedHotelsDatabaseReply{
		Seeded:  result.Seeded,
		Skipped: result.Skipped,
		Total:   result.Total,
		Message: result.Message,
	}, nil
}

// ListHotel keeps backward compatibility with older callers.
func (s *HotelsService) ListHotel(ctx context.Context, req *v1.ListHotelsRequest) (*v1.ListHotelsReply, error) {
	return s.ListHotels(ctx, req)
}

func (s *HotelsService) convertToHotelsList(hotels []*biz.Hotel) []*v1.HotelInfo {
	result := make([]*v1.HotelInfo, len(hotels))
	for i, p := range hotels {
		info, err := s.convertToHotelInfo(p)
		if err != nil {
			s.log.Errorf("Failed to convert hotel %d: %v", p.ID, err)
			continue
		}
		result[i] = info
	}
	return result
}

// =======================HELPER==========================//
func (s *HotelsService) convertToHotelInfo(p *biz.Hotel) (*v1.HotelInfo, error) {
	if p == nil {
		return nil, nil
	}

	discountPct := float64(p.ActualPrice-p.SellingPrice) * 100

	var crawledAt, createdAt, updatedAt *timestamppb.Timestamp
	if !p.CrawledAt.IsZero() {
		crawledAt = timestamppb.New(p.CrawledAt)
	}
	if !p.CreatedAt.IsZero() {
		createdAt = timestamppb.New(p.CreatedAt)
	}
	if !p.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(p.UpdatedAt)
	}

	return &v1.HotelInfo{
		Id:                 p.ID,
		Brand:              p.Brand,
		Description:        p.Description,
		ActualPrice:        p.ActualPrice,
		SellingPrice:       p.SellingPrice,
		Discount:           p.Discount,
		DiscountPercentage: float32(discountPct),
		PriceNumeric:       int32(p.PriceNumeric),
		Category:           p.Category,
		SubCategory:        p.SubCategory,
		OutOfStock:         p.OutOfStock,
		Seller:             p.Seller,
		AverageRating:      p.AverageRating,
		RatingNumeric:      float32(p.RatingNumeric),
		PrimaryImage:       p.Image,
		Url:                p.URL,
		StyleCode:          p.StyleCode,
		CrawledAt:          crawledAt,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		ViewCount:          int32(p.ViewCount),
		ClickCount:         int32(p.ClickCount),
		Featured:           p.Featured,
	}, nil
}
