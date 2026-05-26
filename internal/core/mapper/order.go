package mapper

import (
	"time"

	"github.com/google/uuid"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
)

func DomainToResponse(o *domain.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(o.Items))
	for i, it := range o.Items {
		items[i] = dto.OrderItemResponse{
			ProductID: it.ProductID.String(),
			Name:      it.Name,
			Price:     it.Price,
			Quantity:  it.Quantity,
			Subtotal:  it.Subtotal,
		}
	}
	return dto.OrderResponse{
		ID:          o.ID.String(),
		UserID:      o.UserID.String(),
		Status:      string(o.Status),
		Items:       items,
		TotalAmount: o.TotalAmount,
		DeliveryAddress: dto.DeliveryAddressDTO{
			City:   o.DeliveryAddress.City,
			Street: o.DeliveryAddress.Street,
			Zip:    o.DeliveryAddress.Zip,
		},
		PaymentMethod: o.PaymentMethod,
		CreatedAt:     o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     o.UpdatedAt.Format(time.RFC3339),
	}
}

func CartItemToOrderItem(orderID uuid.UUID, item dto.CartItem) domain.OrderItem {
	pid, _ := uuid.Parse(item.ProductID)
	return domain.OrderItem{
		ID:        uuid.New(),
		OrderID:   orderID,
		ProductID: pid,
		Name:      item.Name,
		Price:     item.Price,
		Quantity:  item.Quantity,
		Subtotal:  item.Price * float64(item.Quantity),
	}
}
