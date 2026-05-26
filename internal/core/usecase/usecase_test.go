package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/port"
	"market-order-service/internal/core/usecase"
)

type stubRepo struct {
	order *domain.Order
	err   error
}

func (s *stubRepo) Create(_ context.Context, o *domain.Order) error {
	s.order = o
	return s.err
}
func (s *stubRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Order, error) {
	return s.order, s.err
}
func (s *stubRepo) ListByUserID(_ context.Context, _ dto.ListOrdersParams) (*dto.ListOrdersResponse, error) {
	return &dto.ListOrdersResponse{}, s.err
}
func (s *stubRepo) UpdateStatus(_ context.Context, _ uuid.UUID, status domain.OrderStatus) error {
	if s.order != nil {
		s.order.Status = status
	}
	return s.err
}

var _ port.OrderRepository = (*stubRepo)(nil)

type stubCart struct {
	items []dto.CartItem
	err   error
}

func (s *stubCart) GetCart(_ context.Context, _ uuid.UUID) (*dto.CartResponse, error) {
	return &dto.CartResponse{Items: s.items}, s.err
}
func (s *stubCart) ClearCart(_ context.Context, _ uuid.UUID) error { return nil }

var _ port.CartClient = (*stubCart)(nil)

func TestCreateOrder_EmptyCart(t *testing.T) {
	uc := usecase.NewCreateOrder(&stubRepo{}, &stubCart{items: []dto.CartItem{}})
	_, err := uc.Execute(context.Background(), usecase.CreateOrderInput{
		UserID:  uuid.New(),
		Request: dto.CreateOrderRequest{PaymentMethod: "card"},
	})
	if err != domain.ErrEmptyCart {
		t.Fatalf("want ErrEmptyCart, got %v", err)
	}
}

func TestUpdateStatus_ForbiddenForBuyer(t *testing.T) {
	order := &domain.Order{ID: uuid.New(), Status: domain.StatusPending}
	uc := usecase.NewUpdateStatus(&stubRepo{order: order})
	_, err := uc.Execute(context.Background(), usecase.UpdateStatusInput{
		OrderID: order.ID,
		Role:    "buyer",
		Request: dto.UpdateStatusRequest{Status: "confirmed"},
	})
	if err != domain.ErrForbidden {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
}

func TestUpdateStatus_InvalidTransition(t *testing.T) {
	order := &domain.Order{ID: uuid.New(), Status: domain.StatusShipped}
	uc := usecase.NewUpdateStatus(&stubRepo{order: order})
	_, err := uc.Execute(context.Background(), usecase.UpdateStatusInput{
		OrderID: order.ID,
		Role:    "seller",
		Request: dto.UpdateStatusRequest{Status: "confirmed"},
	})
	if err != domain.ErrInvalidTransition {
		t.Fatalf("want ErrInvalidTransition, got %v", err)
	}
}

func TestCancelOrder_BuyerCannotCancelConfirmed(t *testing.T) {
	uid := uuid.New()
	order := &domain.Order{ID: uuid.New(), UserID: uid, Status: domain.StatusConfirmed}
	uc := usecase.NewCancelOrder(&stubRepo{order: order})
	_, err := uc.Execute(context.Background(), usecase.CancelOrderInput{
		OrderID: order.ID,
		UserID:  uid,
		Role:    "buyer",
	})
	if err != domain.ErrCancelNotAllowed {
		t.Fatalf("want ErrCancelNotAllowed, got %v", err)
	}
}
