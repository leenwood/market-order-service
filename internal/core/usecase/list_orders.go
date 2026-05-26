package usecase

import (
	"context"

	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/port"
)

type ListOrdersUseCase struct {
	repo port.OrderRepository
}

func NewListOrders(repo port.OrderRepository) *ListOrdersUseCase {
	return &ListOrdersUseCase{repo: repo}
}

func (uc *ListOrdersUseCase) Execute(ctx context.Context, params dto.ListOrdersParams) (*dto.ListOrdersResponse, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}
	return uc.repo.ListByUserID(ctx, params)
}
