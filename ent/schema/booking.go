package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// Booking holds the schema definition for the Booking entity.
type Booking struct {
	ent.Schema
}

// Mixin of the Booking.
func (Booking) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the Booking.
func (Booking) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("hotel_id"),
		field.String("name").NotEmpty(),
		field.Time("booking_date_start"),
		field.Time("booking_date_end"),
		field.Float("price").Min(0),
		field.Int64("number_of_guests").Positive(),
	}
}

// Edges of the Booking.
func (Booking) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("bookings").
			Field("user_id").
			Unique().
			Required(),
		edge.From("hotel", Hotel.Type).
			Ref("bookings").
			Field("hotel_id").
			Unique().
			Required(),
	}
}

// Indexes of the Booking.
func (Booking) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("hotel_id"),
		index.Fields("booking_date_start"),
		index.Fields("booking_date_end"),
	}
}
