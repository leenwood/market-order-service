package usecase

import (
	"context"

	"github.com/google/uuid"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/mapper"
	"market-order-service/internal/core/port"
)

type CancelOrderUseCase struct {
	repo port.OrderRepository
}

func NewCancelOrder(repo port.OrderRepository) *CancelOrderUseCase {
	return &CancelOrderUseCase{repo: repo}
}

type CancelOrderInput struct {
	OrderID uuid.UUID
	UserID  uuid.UUID
	Role    string
}

func (uc *CancelOrderUseCase) Execute(ctx context.Context, in CancelOrderInput) (*dto.OrderResponse, error) {
	order, err := uc.repo.GetByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != in.UserID && in.Role != "seller" && in.Role != "admin" {
		return nil, domain.ErrForbidden
	}

	if !domain.CanTransitionTo(order.Status, domain.StatusCancelled, in.Role) {
		return nil, domain.ErrCancelNotAllowed
	}

	if err := uc.repo.UpdateStatus(ctx, in.OrderID, domain.StatusCancelled); err != nil {
		return nil, err
	}

	order.Status = domain.StatusCancelled
	resp := mapper.DomainToResponse(order)
	return &resp, nil
}
