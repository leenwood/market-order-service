package usecase

import (
	"context"

	"github.com/google/uuid"

	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/mapper"
	"market-order-service/internal/core/port"
)

type GetOrderUseCase struct {
	repo port.OrderRepository
}

func NewGetOrder(repo port.OrderRepository) *GetOrderUseCase {
	return &GetOrderUseCase{repo: repo}
}

func (uc *GetOrderUseCase) Execute(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := mapper.DomainToResponse(order)
	return &resp, nil
}
