package mapper_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/mapper"
)

func TestDomainToResponse(t *testing.T) {
	id := uuid.New()
	uid := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	order := &domain.Order{
		ID:              id,
		UserID:          uid,
		Status:          domain.StatusPending,
		TotalAmount:     299.99,
		PaymentMethod:   "card",
		DeliveryAddress: domain.DeliveryAddress{City: "Moscow", Street: "Pushkin St", Zip: "101000"},
		CreatedAt:       now,
		UpdatedAt:       now,
		Items: []domain.OrderItem{
			{ProductID: uuid.New(), Name: "Item", Price: 299.99, Quantity: 1, Subtotal: 299.99},
		},
	}

	resp := mapper.DomainToResponse(order)
	if resp.ID != id.String() {
		t.Errorf("ID mismatch: want %s got %s", id.String(), resp.ID)
	}
	if resp.Status != "pending" {
		t.Errorf("Status mismatch: want pending got %s", resp.Status)
	}
	if len(resp.Items) != 1 {
		t.Errorf("Items count: want 1 got %d", len(resp.Items))
	}
}
