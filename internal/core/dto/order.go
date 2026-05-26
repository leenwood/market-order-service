package dto

type CreateOrderRequest struct {
	DeliveryAddress DeliveryAddressDTO `json:"delivery_address" binding:"required"`
	PaymentMethod   string             `json:"payment_method" binding:"required,oneof=card cash"`
}

type DeliveryAddressDTO struct {
	City   string `json:"city" binding:"required"`
	Street string `json:"street" binding:"required"`
	Zip    string `json:"zip" binding:"required"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type ListOrdersParams struct {
	UserID   string
	Status   string
	Page     int
	PageSize int
}

type ListOrdersResponse struct {
	Items      []OrderResponse `json:"items"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

type OrderResponse struct {
	ID              string              `json:"id"`
	UserID          string              `json:"user_id"`
	Status          string              `json:"status"`
	Items           []OrderItemResponse `json:"items"`
	TotalAmount     float64             `json:"total_amount"`
	DeliveryAddress DeliveryAddressDTO  `json:"delivery_address"`
	PaymentMethod   string              `json:"payment_method"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
}

type OrderItemResponse struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}
