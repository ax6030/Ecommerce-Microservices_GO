package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jason/ecommerce/order-service/internal/model"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	order.ID = uuid.New().String()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO orders (id, user_id, status, total_amount) VALUES (?, ?, ?, ?)`,
		order.ID, order.UserID, order.Status, order.TotalAmount,
	)
	if err != nil {
		return err
	}

	for i := range order.Items {
		order.Items[i].ID = uuid.New().String()
		order.Items[i].OrderID = order.ID
		_, err = tx.ExecContext(ctx,
			`INSERT INTO order_items (id, order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?, ?)`,
			order.Items[i].ID, order.Items[i].OrderID, order.Items[i].ProductID,
			order.Items[i].Quantity, order.Items[i].UnitPrice,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order
	err := r.db.GetContext(ctx, &order, `SELECT * FROM orders WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	var items []model.OrderItem
	if err := r.db.SelectContext(ctx, &items, `SELECT * FROM order_items WHERE order_id = ?`, id); err != nil {
		return nil, err
	}
	order.Items = items
	return &order, nil
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Order, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM orders WHERE user_id = ?`, userID); err != nil {
		return nil, 0, err
	}
	var orders []model.Order
	if err := r.db.SelectContext(ctx, &orders,
		`SELECT * FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
