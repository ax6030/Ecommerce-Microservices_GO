package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jason/ecommerce/product-service/internal/model"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *model.Product) error {
	p.ID = uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO products (id, name, description, price, stock) VALUES (?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Description, p.Price, p.Stock,
	)
	return err
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*model.Product, error) {
	var p model.Product
	err := r.db.GetContext(ctx, &p, `SELECT * FROM products WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}
	return &p, nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]model.Product, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM products`); err != nil {
		return nil, 0, err
	}
	var products []model.Product
	err := r.db.SelectContext(ctx, &products, `SELECT * FROM products ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	return products, total, err
}

func (r *ProductRepository) DeductStock(ctx context.Context, id string, qty int32) (int32, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var stock int32
	if err := tx.GetContext(ctx, &stock, `SELECT stock FROM products WHERE id = ? FOR UPDATE`, id); err != nil {
		return 0, fmt.Errorf("product not found: %w", err)
	}
	if stock < qty {
		return 0, fmt.Errorf("insufficient stock: have %d, need %d", stock, qty)
	}

	newStock := stock - qty
	if _, err := tx.ExecContext(ctx, `UPDATE products SET stock = ? WHERE id = ?`, newStock, id); err != nil {
		return 0, err
	}
	return newStock, tx.Commit()
}
