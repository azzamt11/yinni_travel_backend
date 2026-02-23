package service

import (
	"context"
	"time"

	v1 "yinni_travel_backend/api/booking/v1"
	"yinni_travel_backend/app/booking/internal/biz"
	"yinni_travel_backend/pkg/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BookingService struct {
	v1.UnimplementedBookingServer

	uc  *biz.BookingUsecase
	log *log.Helper
}

func NewBookingService(uc *biz.BookingUsecase) *BookingService {
	return &BookingService{uc: uc}
}

func (s *BookingService) CreateBooking(ctx context.Context, req *v1.CreateBookingRequest) (*v1.CreateBookingReply, error) {
	userID := req.GetUserId()
	if userID <= 0 {
		if claims, err := middleware.ExtractClaimsFromContext(ctx); err == nil {
			userID = claims.UserID
		}
	}

	booking, err := s.uc.CreateBooking(ctx, &biz.CreateBookingParams{
		UserID:           userID,
		HotelID:          req.GetHotelId(),
		Name:             req.GetName(),
		BookingDateStart: asTime(req.GetBookingDateStart()),
		BookingDateEnd:   asTime(req.GetBookingDateEnd()),
		Price:            req.GetPrice(),
		NumberOfGuests:   req.GetNumberOfGuests(),
	})
	if err != nil {
		return nil, err
	}

	return &v1.CreateBookingReply{
		Booking: toProtoBooking(booking),
	}, nil
}
func (s *BookingService) SeeBookings(ctx context.Context, req *v1.SeeBookingsRequest) (*v1.SeeBookingsReply, error) {
	userID := req.GetUserId()
	if userID <= 0 {
		if claims, err := middleware.ExtractClaimsFromContext(ctx); err == nil {
			userID = claims.UserID
		}
	}

	bookings, err := s.uc.SeeBookings(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.BookingInfo, 0, len(bookings))
	for _, booking := range bookings {
		items = append(items, toProtoBooking(booking))
	}

	return &v1.SeeBookingsReply{
		Bookings: items,
	}, nil
}

func toProtoBooking(in *biz.Booking) *v1.BookingInfo {
	if in == nil {
		return nil
	}
	return &v1.BookingInfo{
		Id:               in.ID,
		UserId:           in.UserID,
		HotelId:          in.HotelID,
		Name:             in.Name,
		BookingDateStart: timestamppb.New(in.BookingDateStart),
		BookingDateEnd:   timestamppb.New(in.BookingDateEnd),
		Price:            in.Price,
		NumberOfGuests:   in.NumberOfGuests,
		CreatedAt:        timestamppb.New(in.CreatedAt),
		UpdatedAt:        timestamppb.New(in.UpdatedAt),
	}
}

func asTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
