package dto

type CartResponse struct {
	Items []CartItem `json:"items"`
}

type CartItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}
