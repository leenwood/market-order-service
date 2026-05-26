package port

import (
	"context"

	"github.com/google/uuid"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	ListByUserID(ctx context.Context, params dto.ListOrdersParams) (*dto.ListOrdersResponse, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
}

type CartClient interface {
	GetCart(ctx context.Context, userID uuid.UUID) (*dto.CartResponse, error)
	ClearCart(ctx context.Context, userID uuid.UUID) error
}
