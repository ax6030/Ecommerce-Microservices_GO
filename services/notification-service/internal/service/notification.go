package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/jason/ecommerce/notification-service/internal/model"
	"github.com/jason/ecommerce/notification-service/internal/messaging"
)

type NotificationService struct {
	consumer *messaging.Consumer
	logger   *slog.Logger
}

func NewNotificationService(consumer *messaging.Consumer) *NotificationService {
	return &NotificationService{
		consumer: consumer,
		logger:   slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "notification-service"),
	}
}

func (s *NotificationService) Subscribe() error {
	deliveries, err := s.consumer.Consume()
	if err != nil {
		return err
	}
	go s.processDeliveries(deliveries)
	return nil
}

func (s *NotificationService) processDeliveries(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		var event model.OrderCreatedEvent
		if err := json.Unmarshal(d.Body, &event); err != nil {
			s.logger.Error("failed to unmarshal order.created event", "error", err)
			d.Nack(false, false)
			continue
		}
		s.handleOrderCreated(event)
		d.Ack(false)
	}
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
