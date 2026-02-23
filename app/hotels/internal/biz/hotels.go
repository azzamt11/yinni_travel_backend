package biz

import (
	"context"
	"time"

	v1 "yinni_travel_backend/api/hotels/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	ErrHotelNotFound     = errors.NotFound(v1.ErrorReason_HOTEL_NOT_FOUND.String(), "user not found")
	ErrInvalidPriceRange = errors.BadRequest(v1.ErrorReason_INVALID_PRICE_RANGE.String(), "invalid price range")
)

type Hotel struct {
	ID             int64
	Name           string
	Brand          string
	Description    string
	ActualPrice    float64
	SellingPrice   float64
	Discount       string
	PriceNumeric   int
	Category       string
	SubCategory    string
	OutOfStock     bool
	Seller         string
	AverageRating  string
	RatingNumeric  float32
	Image          string
	HotelDetails   []map[string]string
	URL            string
	PID            string
	StyleCode      string
	CrawledAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ViewCount      int
	ClickCount     int
	Featured       bool
	Embedding      []float32
	SearchKeywords []string
}

type ListHotelsParams struct {
	Page        int32
	PageSize    int32
	Category    string
	Brand       string
	SubCategory string
	MinPrice    int32
	MaxPrice    int32
	MinRating   float32
	InStock     bool
	Featured    bool
	Seller      string
	SortBy      string
	SortOrder   string
	SearchQuery string
}

type HotelsRepo interface {
	GetHotelsList(context.Context, *ListHotelsParams) ([]*Hotel, int64, error)
	SeedHotelsDatabase(context.Context, *SeedHotelsParams) (*SeedHotelsResult, error)
}

type HotelsUsecase struct {
	repo HotelsRepo
	log  *log.Helper
}

func NewHotelsUsecase(repo HotelsRepo) *HotelsUsecase {
	return &HotelsUsecase{repo: repo}
}

func (uc *HotelsUsecase) GetHotelsList(ctx context.Context, params *ListHotelsParams) ([]*Hotel, int64, error) {
	uc.log.Infof("ListHotels: page=%d, pageSize=%d", params.Page, params.PageSize)

	if err := params.Validate(); err != nil {
		return nil, 0, err
	}

	return uc.repo.GetHotelsList(ctx, params)
}

type SeedHotelsParams struct {
	ClearExisting bool
	MaxHotels     int32
	DatasetPath   string
}

type SeedHotelsResult struct {
	Seeded  int32
	Skipped int32
	Total   int32
	Message string
}

func (uc *HotelsUsecase) SeedHotelsDatabase(ctx context.Context, params *SeedHotelsParams) (*SeedHotelsResult, error) {
	if params == nil {
		params = &SeedHotelsParams{}
	}
	if params.MaxHotels < 0 {
		params.MaxHotels = 0
	}
	return uc.repo.SeedHotelsDatabase(ctx, params)
}

func (p *ListHotelsParams) Validate() error {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	if p.MinPrice < 0 {
		p.MinPrice = 0
	}
	if p.MaxPrice < 0 {
		p.MaxPrice = 0
	}
	if p.MinPrice > p.MaxPrice && p.MaxPrice > 0 {
		return ErrInvalidPriceRange
	}
	return nil
}
