package biz

import (
	"context"
	"time"

	v1 "yinni-travel-backend/api/booking/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	ErrBookingNotFound   = errors.NotFound(v1.ErrorReason_BOOKING_NOT_FOUND.String(), "booking not found")
	ErrInvalidBookingID  = errors.BadRequest(v1.ErrorReason_INVALID_BOOKING_ID.String(), "invalid booking id")
	ErrInvalidUserID     = errors.BadRequest(v1.ErrorReason_INVALID_USER_ID.String(), "invalid user id")
	ErrInvalidHotelID    = errors.BadRequest(v1.ErrorReason_INVALID_HOTEL_ID.String(), "invalid hotel id")
	ErrInvalidBookingDay = errors.BadRequest(v1.ErrorReason_INVALID_BOOKING_DATE.String(), "invalid booking date")
	ErrInvalidParams     = errors.BadRequest(v1.ErrorReason_INVALID_PARAMETERS.String(), "invalid parameters")
	ErrBookingCreateFail = errors.InternalServer(v1.ErrorReason_BOOKING_CREATION_FAILED.String(), "booking creation failed")
)

type Booking struct {
	ID               int64
	UserID           int64
	HotelID          int64
	Name             string
	BookingDateStart time.Time
	BookingDateEnd   time.Time
	Price            float64
	NumberOfGuests   int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CreateBookingParams struct {
	UserID           int64
	HotelID          int64
	Name             string
	BookingDateStart time.Time
	BookingDateEnd   time.Time
	Price            float64
	NumberOfGuests   int64
}

type BookingRepo interface {
	CreateBooking(ctx context.Context, params *CreateBookingParams) (*Booking, error)
	SeeBookings(ctx context.Context, userID int64) ([]*Booking, error)
}

type BookingUsecase struct {
	repo BookingRepo
	log  *log.Helper
}

func NewBookingUsecase(repo BookingRepo, logger log.Logger) *BookingUsecase {
	return &BookingUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *BookingUsecase) CreateBooking(ctx context.Context, params *CreateBookingParams) (*Booking, error) {
	if params == nil {
		return nil, ErrInvalidParams
	}
	if params.UserID <= 0 {
		return nil, ErrInvalidUserID
	}
	if params.HotelID <= 0 {
		return nil, ErrInvalidHotelID
	}
	if params.BookingDateStart.IsZero() || params.BookingDateEnd.IsZero() {
		return nil, ErrInvalidBookingDay
	}
	if params.BookingDateEnd.Before(params.BookingDateStart) {
		return nil, ErrInvalidBookingDay
	}
	if params.NumberOfGuests <= 0 {
		return nil, ErrInvalidParams
	}
	if params.Price < 0 {
		return nil, ErrInvalidParams
	}
	if params.Name == "" {
		return nil, ErrInvalidParams
	}

	uc.log.WithContext(ctx).Infof("creating booking for user=%d hotel=%d", params.UserID, params.HotelID)
	booking, err := uc.repo.CreateBooking(ctx, params)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("create booking failed: %v", err)
		return nil, ErrBookingCreateFail
	}
	return booking, nil
}

func (uc *BookingUsecase) SeeBookings(ctx context.Context, userID int64) ([]*Booking, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	uc.log.WithContext(ctx).Infof("loading bookings for user=%d", userID)
	return uc.repo.SeeBookings(ctx, userID)
}
