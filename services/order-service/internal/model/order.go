package model

import "time"

type Order struct {
	ID          string      `db:"id"`
	UserID      string      `db:"user_id"`
	Status      string      `db:"status"`
	TotalAmount float64     `db:"total_amount"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
	Items       []OrderItem `db:"-"`
}

type OrderItem struct {
	ID        string  `db:"id"`
	OrderID   string  `db:"order_id"`
	ProductID string  `db:"product_id"`
	Quantity  int32   `db:"quantity"`
	UnitPrice float64 `db:"unit_price"`
}

type OrderCreatedEvent struct {
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
}
