package data

import (
	"context"
	"time"

	"yinni_travel_backend/app/booking/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type bookingRepo struct {
	data *Data
	log  *log.Helper
}

// NewBookingRepo .
func NewBookingRepo(data *Data, logger log.Logger) biz.BookingRepo {
	return &bookingRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *bookingRepo) CreateBooking(ctx context.Context, params *biz.CreateBookingParams) (*biz.Booking, error) {
	now := time.Now()
	start := params.BookingDateStart.UTC()
	end := params.BookingDateEnd.UTC()

	r.data.mu.Lock()
	defer r.data.mu.Unlock()

	id := r.data.nextBookingID
	r.data.nextBookingID++

	record := &BookingRecord{
		ID:               id,
		UserID:           params.UserID,
		HotelID:          params.HotelID,
		Name:             params.Name,
		BookingDateStart: start.Unix(),
		BookingDateEnd:   end.Unix(),
		Price:            params.Price,
		NumberOfGuests:   params.NumberOfGuests,
		CreatedAtUnix:    now.Unix(),
		UpdatedAtUnix:    now.Unix(),
	}
	r.data.bookings[params.UserID] = append(r.data.bookings[params.UserID], record)

	return toBizBooking(record), nil
}

func (r *bookingRepo) SeeBookings(ctx context.Context, userID int64) ([]*biz.Booking, error) {
	r.data.mu.RLock()
	defer r.data.mu.RUnlock()

	rows := r.data.bookings[userID]
	result := make([]*biz.Booking, 0, len(rows))
	for _, row := range rows {
		result = append(result, toBizBooking(row))
	}
	return result, nil
}

func toBizBooking(r *BookingRecord) *biz.Booking {
	return &biz.Booking{
		ID:               r.ID,
		UserID:           r.UserID,
		HotelID:          r.HotelID,
		Name:             r.Name,
		BookingDateStart: time.Unix(r.BookingDateStart, 0).UTC(),
		BookingDateEnd:   time.Unix(r.BookingDateEnd, 0).UTC(),
		Price:            r.Price,
		NumberOfGuests:   r.NumberOfGuests,
		CreatedAt:        time.Unix(r.CreatedAtUnix, 0).UTC(),
		UpdatedAt:        time.Unix(r.UpdatedAtUnix, 0).UTC(),
	}
}
