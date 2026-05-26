package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Status          OrderStatus
	Items           []OrderItem
	TotalAmount     float64
	PaymentMethod   string
	DeliveryAddress DeliveryAddress
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Name      string
	Price     float64
	Quantity  int
	Subtotal  float64
}

type DeliveryAddress struct {
	City   string
	Street string
	Zip    string
}
