package data

import (
	"sync"
	"yinni-travel-backend/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewBookingRepo)

// Data .
type Data struct {
	mu            sync.RWMutex
	nextBookingID int64
	bookings      map[int64][]*BookingRecord
}

type BookingRecord struct {
	ID               int64
	UserID           int64
	HotelID          int64
	Name             string
	BookingDateStart int64
	BookingDateEnd   int64
	Price            float64
	NumberOfGuests   int64
	CreatedAtUnix    int64
	UpdatedAtUnix    int64
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
	}
	return &Data{
		nextBookingID: 1,
		bookings:      make(map[int64][]*BookingRecord),
	}, cleanup, nil
}
