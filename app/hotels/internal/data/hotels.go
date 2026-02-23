package data

import (
	"context"
	"strings"
	"trivgoo-backend/app/hotels/internal/biz"
	"trivgoo-backend/ent"
	"trivgoo-backend/ent/hotel"

	"github.com/go-kratos/kratos/v2/log"
)

type hotelsRepo struct {
	data *Data
	log  *log.Helper
}

// NewHotelRepo .
func NewHotelsRepo(data *Data, logger log.Logger) biz.HotelsRepo {
	return &hotelsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *hotelsRepo) GetHotelsList(ctx context.Context, params *biz.ListHotelsParams) ([]*biz.Hotel, int64, error) {
	query := r.data.ent.Hotel.Query()

	// Apply filters
	if params.Category != "" {
		query = query.Where(hotel.Category(params.Category))
	}
	if params.SubCategory != "" {
		query = query.Where(hotel.SubCategory(params.SubCategory))
	}
	if params.Brand != "" {
		query = query.Where(hotel.Brand(params.Brand))
	}
	if params.Seller != "" {
		query = query.Where(hotel.Seller(params.Seller))
	}
	if params.MinPrice > 0 {
		query = query.Where(hotel.PriceNumericGTE(int(params.MinPrice)))
	}
	if params.MaxPrice > 0 {
		query = query.Where(hotel.PriceNumericLTE(int(params.MaxPrice)))
	}
	if params.MinRating > 0 {
		query = query.Where(hotel.RatingNumericGTE(float64(params.MinRating)))
	}
	if params.InStock {
		query = query.Where(hotel.OutOfStock(false))
	}
	if params.Featured {
		query = query.Where(hotel.Featured(true))
	}

	// Apply search query if provided
	if params.SearchQuery != "" {
		// Simple text search (could be enhanced with full-text search)
		query = query.Where(
			hotel.Or(
				hotel.TitleContains(params.SearchQuery),
				hotel.DescriptionContains(params.SearchQuery),
				hotel.BrandContains(params.SearchQuery),
			),
		)
	}

	// Apply sorting
	switch strings.ToLower(params.SortBy) {
	case "price":
		if strings.ToLower(params.SortOrder) == "asc" {
			query = query.Order(ent.Asc(hotel.FieldPriceNumeric))
		} else {
			query = query.Order(ent.Desc(hotel.FieldPriceNumeric))
		}
	case "rating":
		if strings.ToLower(params.SortOrder) == "asc" {
			query = query.Order(ent.Asc(hotel.FieldRatingNumeric))
		} else {
			query = query.Order(ent.Desc(hotel.FieldRatingNumeric))
		}
	case "newest":
		query = query.Order(ent.Desc(hotel.FieldCreateTime))
	case "popular":
		query = query.Order(ent.Desc(hotel.FieldViewCount))
	default:
		query = query.Order(ent.Desc(hotel.FieldCreateTime))
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (int(params.Page) - 1) * int(params.PageSize)
	rows, err := query.
		Offset(offset).
		Limit(int(params.PageSize)).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	rv := make([]*biz.Hotel, 0, len(rows))
	for _, row := range rows {
		rv = append(rv, &biz.Hotel{
			ID:             int64(row.ID),
			Name:           row.Title,
			Brand:          row.Brand,
			Description:    row.Description,
			ActualPrice:    float64(row.ActualPrice),
			SellingPrice:   float64(row.SellingPrice),
			Discount:       row.Discount,
			PriceNumeric:   row.PriceNumeric,
			Category:       row.Category,
			SubCategory:    row.SubCategory,
			OutOfStock:     row.OutOfStock,
			Seller:         row.Seller,
			AverageRating:  row.AverageRating,
			RatingNumeric:  float32(row.RatingNumeric),
			Image:          row.Image,
			URL:            row.URL,
			CrawledAt:      row.CrawledAt,
			CreatedAt:      row.CreateTime,
			UpdatedAt:      row.UpdateTime,
			ViewCount:      row.ViewCount,
			ClickCount:     row.ClickCount,
			Featured:       row.Featured,
			SearchKeywords: row.SearchKeywords,
		})
	}
	return rv, int64(total), nil
}
