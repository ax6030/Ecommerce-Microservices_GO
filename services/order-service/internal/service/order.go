package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"
	pb_product "github.com/jason/ecommerce/proto/product"
	"github.com/jason/ecommerce/order-service/internal/model"
	"github.com/jason/ecommerce/order-service/internal/repository"
)

type OrderService struct {
	repo        *repository.OrderRepository
	productConn pb_product.ProductServiceClient
	nc          *nats.Conn
	logger      *slog.Logger
}

func NewOrderService(
	repo *repository.OrderRepository,
	productConn pb_product.ProductServiceClient,
	nc *nats.Conn,
) *OrderService {
	return &OrderService{
		repo:        repo,
		productConn: productConn,
		nc:          nc,
		logger:      slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "order-service"),
	}
}

type CreateOrderInput struct {
	UserID string
	Items  []ItemInput
}

type ItemInput struct {
	ProductID string
	Quantity  int32
}

func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*model.Order, error) {
	var orderItems []model.OrderItem
	totalAmount := 0.0

	for _, item := range input.Items {
		product, err := s.productConn.GetProduct(ctx, &pb_product.GetProductRequest{Id: item.ProductID})
		if err != nil {
			return nil, fmt.Errorf("product %s not found: %w", item.ProductID, err)
		}

		_, err = s.productConn.DeductStock(ctx, &pb_product.DeductStockRequest{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to deduct stock for %s: %w", item.ProductID, err)
		}

		orderItems = append(orderItems, model.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: product.Price,
		})
		totalAmount += product.Price * float64(item.Quantity)
	}

	order := &model.Order{
		UserID:      input.UserID,
		Status:      "confirmed",
		TotalAmount: totalAmount,
		Items:       orderItems,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	s.publishOrderCreated(order)
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OrderService) ListOrders(ctx context.Context, userID string, page, limit int) ([]model.Order, int, error) {
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListByUserID(ctx, userID, limit, offset)
}

func (s *OrderService) publishOrderCreated(order *model.Order) {
	event := model.OrderCreatedEvent{
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Items:       order.Items,
	}
	data, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("failed to marshal order event", "error", err)
		return
	}
	if err := s.nc.Publish("order.created", data); err != nil {
		s.logger.Error("failed to publish order.created event", "error", err)
		return
	}
	s.logger.Info("published order.created event", "order_id", order.ID)
}
