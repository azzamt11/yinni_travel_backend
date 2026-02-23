package data

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"yinni-travel-backend/app/hotels/internal/biz"
	"yinni-travel-backend/ent"
	"yinni-travel-backend/ent/hotel"

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

func (r *hotelsRepo) SeedHotelsDatabase(ctx context.Context, params *biz.SeedHotelsParams) (*biz.SeedHotelsResult, error) {
	datasetPath := strings.TrimSpace(params.DatasetPath)
	if datasetPath == "" {
		datasetPath = resolveDatasetPath()
	}

	raw, err := os.ReadFile(datasetPath)
	if err != nil {
		return nil, fmt.Errorf("read dataset %q: %w", datasetPath, err)
	}

	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse dataset json: %w", err)
	}

	if params.ClearExisting {
		if _, err := r.data.ent.Hotel.Delete().Exec(ctx); err != nil {
			return nil, fmt.Errorf("clear hotels: %w", err)
		}
	}

	total := len(rows)
	limit := total
	if params.MaxHotels > 0 && int(params.MaxHotels) < total {
		limit = int(params.MaxHotels)
	}

	var seeded, skipped int32
	for i := 0; i < limit; i++ {
		item := rows[i]
		origID := getString(item, "_id", "original_id", "id")
		pid := getString(item, "pid", "product_id", "hotel_id")
		title := getString(item, "title", "name")
		brand := getString(item, "brand")
		category := getString(item, "category")
		subCategory := getString(item, "sub_category", "subcategory")

		if origID == "" && pid == "" {
			origID = fmt.Sprintf("row-%d", i+1)
			pid = origID
		}
		if origID == "" {
			origID = pid
		}
		if pid == "" {
			pid = origID
		}
		if title == "" {
			title = "Untitled Hotel"
		}
		if brand == "" {
			brand = "Unknown"
		}
		if category == "" {
			category = "uncategorized"
		}
		if subCategory == "" {
			subCategory = "general"
		}

		exists, err := r.data.ent.Hotel.Query().
			Where(hotel.Or(hotel.OriginalIDEQ(origID), hotel.PidEQ(pid))).
			Exist(ctx)
		if err != nil {
			return nil, err
		}
		if exists {
			skipped++
			continue
		}

		actualPrice := float32(getFloat(item, "actual_price", "actualPrice", "mrp"))
		sellingPrice := float32(getFloat(item, "selling_price", "sellingPrice", "price"))
		if sellingPrice == 0 && actualPrice > 0 {
			sellingPrice = actualPrice
		}
		priceNumeric := int(getFloat(item, "price_numeric"))
		if priceNumeric == 0 {
			priceNumeric = int(sellingPrice)
		}
		ratingNumeric := getFloat(item, "rating_numeric")
		if ratingNumeric == 0 {
			ratingNumeric = getFloat(item, "rating")
		}

		create := r.data.ent.Hotel.Create().
			SetOriginalID(origID).
			SetPid(pid).
			SetTitle(title).
			SetBrand(brand).
			SetCategory(category).
			SetSubCategory(subCategory).
			SetDescription(getString(item, "description")).
			SetActualPrice(actualPrice).
			SetSellingPrice(sellingPrice).
			SetDiscount(getString(item, "discount")).
			SetOutOfStock(getBool(item, "out_of_stock", "outOfStock")).
			SetSeller(getString(item, "seller")).
			SetAverageRating(getString(item, "average_rating", "averageRating")).
			SetImage(getString(item, "image", "primary_image")).
			SetURL(getString(item, "url")).
			SetStyleCode(getString(item, "style_code", "styleCode")).
			SetPriceNumeric(priceNumeric).
			SetRatingNumeric(ratingNumeric).
			SetFeatured(getBool(item, "featured")).
			SetViewCount(int(getFloat(item, "view_count", "viewCount"))).
			SetClickCount(int(getFloat(item, "click_count", "clickCount"))).
			SetSearchKeywords(getStringSlice(item, "search_keywords")).
			SetCrawledAt(time.Now())

		if details := getDetails(item, "hotel_details"); len(details) > 0 {
			create = create.SetHotelDetails(details)
		}
		if emb := getFloat32Slice(item, "embedding"); len(emb) > 0 {
			create = create.SetEmbedding(emb)
		}

		if _, err := create.Save(ctx); err != nil {
			// skip invalid/duplicate rows and continue seeding
			r.log.WithContext(ctx).Warnf("skip row %d (%s): %v", i+1, pid, err)
			skipped++
			continue
		}
		seeded++
	}

	return &biz.SeedHotelsResult{
		Seeded:  seeded,
		Skipped: skipped,
		Total:   int32(limit),
		Message: fmt.Sprintf("seed complete from %s", datasetPath),
	}, nil
}

func resolveDatasetPath() string {
	if p := os.Getenv("HOTELS_DATASET_PATH"); strings.TrimSpace(p) != "" {
		return p
	}
	candidates := []string{
		"/app/hotels_dataset.json",
		"./hotels_dataset.json",
		"app/hotels/hotels_dataset.json",
		filepath.Join("..", "..", "hotels_dataset.json"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "app/hotels/hotels_dataset.json"
}

func getString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if s := strings.TrimSpace(t); s != "" {
				return s
			}
		case json.Number:
			return t.String()
		case float64:
			return strconv.FormatFloat(t, 'f', -1, 64)
		case int:
			return strconv.Itoa(t)
		}
	}
	return ""
}

var nonNumber = regexp.MustCompile(`[^0-9.\-]+`)

func getFloat(m map[string]any, keys ...string) float64 {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case float64:
			return t
		case float32:
			return float64(t)
		case int:
			return float64(t)
		case int64:
			return float64(t)
		case json.Number:
			f, _ := t.Float64()
			return f
		case string:
			clean := nonNumber.ReplaceAllString(t, "")
			if clean == "" {
				continue
			}
			f, err := strconv.ParseFloat(clean, 64)
			if err == nil {
				return f
			}
		}
	}
	return 0
}

func getBool(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case bool:
			return t
		case string:
			return strings.EqualFold(t, "true") || t == "1" || strings.EqualFold(t, "yes")
		case float64:
			return t != 0
		}
	}
	return false
}

func getStringSlice(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func getFloat32Slice(m map[string]any, key string) []float32 {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]float32, 0, len(arr))
	for _, item := range arr {
		switch t := item.(type) {
		case float64:
			out = append(out, float32(t))
		case float32:
			out = append(out, t)
		}
	}
	return out
}

func getDetails(m map[string]any, key string) []map[string]string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]string, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		flat := make(map[string]string, len(obj))
		for k, val := range obj {
			flat[k] = fmt.Sprintf("%v", val)
		}
		result = append(result, flat)
	}
	return result
}
