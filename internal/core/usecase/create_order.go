package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/mapper"
	"market-order-service/internal/core/port"
)

type CreateOrderUseCase struct {
	repo port.OrderRepository
	cart port.CartClient
}

func NewCreateOrder(repo port.OrderRepository, cart port.CartClient) *CreateOrderUseCase {
	return &CreateOrderUseCase{repo: repo, cart: cart}
}

type CreateOrderInput struct {
	UserID  uuid.UUID
	Request dto.CreateOrderRequest
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, in CreateOrderInput) (*dto.OrderResponse, error) {
	cartResp, err := uc.cart.GetCart(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if len(cartResp.Items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	orderID := uuid.New()
	items := make([]domain.OrderItem, len(cartResp.Items))
	var total float64
	for i, ci := range cartResp.Items {
		item := mapper.CartItemToOrderItem(orderID, ci)
		total += item.Subtotal
		items[i] = item
	}

	now := time.Now().UTC()
	order := &domain.Order{
		ID:     orderID,
		UserID: in.UserID,
		Status: domain.StatusPending,
		Items:  items,
		DeliveryAddress: domain.DeliveryAddress{
			City:   in.Request.DeliveryAddress.City,
			Street: in.Request.DeliveryAddress.Street,
			Zip:    in.Request.DeliveryAddress.Zip,
		},
		TotalAmount:   total,
		PaymentMethod: in.Request.PaymentMethod,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	_ = uc.cart.ClearCart(ctx, in.UserID)

	resp := mapper.DomainToResponse(order)
	return &resp, nil
}
