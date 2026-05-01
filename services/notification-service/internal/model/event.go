package model

type OrderItem struct {
	OrderID   string  `json:"order_id"`
	ProductID string  `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type OrderCreatedEvent struct {
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
}
