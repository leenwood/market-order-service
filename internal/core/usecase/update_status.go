package usecase

import (
	"context"

	"github.com/google/uuid"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/mapper"
	"market-order-service/internal/core/port"
)

type UpdateStatusUseCase struct {
	repo port.OrderRepository
}

func NewUpdateStatus(repo port.OrderRepository) *UpdateStatusUseCase {
	return &UpdateStatusUseCase{repo: repo}
}

type UpdateStatusInput struct {
	OrderID uuid.UUID
	Role    string
	Request dto.UpdateStatusRequest
}

func (uc *UpdateStatusUseCase) Execute(ctx context.Context, in UpdateStatusInput) (*dto.OrderResponse, error) {
	if in.Role != "seller" && in.Role != "admin" {
		return nil, domain.ErrForbidden
	}

	order, err := uc.repo.GetByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}

	newStatus := domain.OrderStatus(in.Request.Status)
	if !domain.CanTransitionTo(order.Status, newStatus, in.Role) {
		return nil, domain.ErrInvalidTransition
	}

	if err := uc.repo.UpdateStatus(ctx, in.OrderID, newStatus); err != nil {
		return nil, err
	}

	order.Status = newStatus
	resp := mapper.DomainToResponse(order)
	return &resp, nil
}
