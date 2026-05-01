package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/jason/ecommerce/notification-service/internal/model"
)

type NotificationService struct {
	nc     *nats.Conn
	logger *slog.Logger
}

func NewNotificationService(nc *nats.Conn) *NotificationService {
	return &NotificationService{
		nc:     nc,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "notification-service"),
	}
}

func (s *NotificationService) Subscribe() error {
	_, err := s.nc.Subscribe("order.created", func(msg *nats.Msg) {
		var event model.OrderCreatedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			s.logger.Error("failed to unmarshal order.created event", "error", err)
			return
		}
		s.handleOrderCreated(event)
	})
	return err
}

func (s *NotificationService) handleOrderCreated(event model.OrderCreatedEvent) {
	s.logger.Info("received order.created event",
		"order_id", event.OrderID,
		"user_id", event.UserID,
		"total_amount", event.TotalAmount,
		"item_count", len(event.Items),
	)

	notification := fmt.Sprintf(
		"[EMAIL] To: user_%s@example.com | Subject: Order Confirmed | Order #%s confirmed. Total: $%.2f",
		event.UserID, event.OrderID, event.TotalAmount,
	)
	s.logger.Info("sending notification", "notification", notification)
}
