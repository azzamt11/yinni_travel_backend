package schema

import (
	"strconv"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// Hotel holds the schema definition for the Hotel entity.
type Hotel struct {
	ent.Schema
}

// Mixins for Hotel
func (Hotel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the Hotel.
func (Hotel) Fields() []ent.Field {
	return []ent.Field{
		// Original ID from dataset
		field.String("original_id").
			Unique().
			Comment("Original _id from dataset"),

		// Basic hotel info
		field.String("title").
			NotEmpty().
			Comment("Hotel title"),
		field.String("brand").
			NotEmpty().
			Comment("Brand name"),
		field.Text("description").
			Optional().
			Comment("Hotel description"),

		// Pricing
		field.Float32("actual_price").
			Optional().
			Comment("Original price"),
		field.Float32("selling_price").
			Optional().
			Comment("Discounted price"),
		field.String("discount").
			Optional().
			Comment("Discount percentage"),

		// Categorization
		field.String("category").
			NotEmpty().
			Comment("Main category"),
		field.String("sub_category").
			NotEmpty().
			Comment("Sub category"),

		// Stock & Seller
		field.Bool("out_of_stock").
			Default(false),
		field.String("seller").
			Optional().
			Comment("Seller name"),

		// Ratings
		field.String("average_rating").
			Optional().
			Comment("Average rating"),

		// Images
		field.String("image").
			Optional().
			Comment("Hotel image URLs"),

		// Hotel details (dynamic JSON)
		field.JSON("hotel_details", []map[string]string{}).
			Optional().
			Comment("Dynamic hotel attributes"),

		// URLs and identifiers
		field.String("url").
			Optional().
			MaxLen(2048).
			Comment("Hotel URL"),
		field.String("pid").
			Unique().
			Comment("Flipkart hotel ID"),
		field.String("style_code").
			Optional().
			Comment("Style/sku code"),

		// Crawled timestamp
		field.Time("crawled_at").
			Optional().
			Comment("When the data was crawled"),

		// Embeddings for vector search
		field.JSON("embedding", []float32{}).
			Optional().
			Comment("Vector embeddings for semantic search"),

		// Search index fields
		field.JSON("search_keywords", []string{}).
			Optional().
			Comment("Keywords for full-text search"),

		// Metadata
		field.Bool("featured").
			Default(false).
			Comment("Featured hotel"),
		field.Int("view_count").
			Default(0).
			Comment("Number of views"),
		field.Int("click_count").
			Default(0).
			Comment("Number of clicks"),

		// Price as integer for sorting/filtering
		field.Int("price_numeric").
			Optional().
			Min(0).
			Comment("Price as integer for filtering"),

		// Rating as float for sorting
		field.Float("rating_numeric").
			Optional().
			Min(0).
			Max(5).
			Comment("Rating as float for sorting"),
	}
}

// Edges of the Hotel.
func (Hotel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bookings", Booking.Type),
	}
}

// Indexes for Hotel
func (Hotel) Indexes() []ent.Index {
	return []ent.Index{
		// Primary indexes
		index.Fields("pid").Unique(),
		index.Fields("original_id").Unique(),

		// Search indexes
		index.Fields("brand"),
		index.Fields("category"),
		index.Fields("sub_category"),

		// Performance indexes
		index.Fields("price_numeric"),
		index.Fields("rating_numeric"),
		index.Fields("out_of_stock"),
		index.Fields("featured"),

		// Composite indexes for common queries
		index.Fields("category", "sub_category"),
		index.Fields("brand", "category"),
		index.Fields("category", "price_numeric"),
		index.Fields("category", "rating_numeric"),

		// Full-text search (if supported by your DB)
		// index.Fields("title", "description", "search_keywords").StorageKey("hotel_search_idx"),
	}
}

// Helper functions (add these in a separate helper file)
func extractPriceNumber(priceStr string) int {
	// Remove commas, ₹, $, etc. and convert to integer
	priceStr = strings.ReplaceAll(priceStr, ",", "")
	priceStr = strings.ReplaceAll(priceStr, "₹", "")
	priceStr = strings.ReplaceAll(priceStr, "$", "")
	priceStr = strings.ReplaceAll(priceStr, " ", "")

	if price, err := strconv.Atoi(priceStr); err == nil {
		return price
	}
	return 0
}

func extractRatingNumber(ratingStr string) float64 {
	if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil {
		return rating
	}
	return 0.0
}

func generateKeywords(text string) []string {
	words := strings.Fields(strings.ToLower(text))

	// Common words to exclude
	stopWords := map[string]bool{
		"and": true, "or": true, "the": true, "a": true, "an": true,
		"in": true, "on": true, "at": true, "to": true, "for": true,
		"with": true, "by": true, "of": true, "men": true, "women": true,
	}

	var keywords []string
	for _, word := range words {
		if !stopWords[word] && len(word) > 2 {
			keywords = append(keywords, word)
		}
	}

	return keywords
}
