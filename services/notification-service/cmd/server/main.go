package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jason/ecommerce/notification-service/internal/messaging"
	"github.com/jason/ecommerce/notification-service/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "notification-service")

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	consumer, err := messaging.NewConsumer(rabbitmqURL)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()
	logger.Info("connected to RabbitMQ", "url", rabbitmqURL)

	svc := service.NewNotificationService(consumer)
	if err := svc.Subscribe(); err != nil {
		logger.Error("failed to subscribe", "error", err)
		os.Exit(1)
	}
	logger.Info("subscribed to order.created queue")

	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"notification-service"}`))
		})
		logger.Info("health server listening", "port", "8084")
		http.ListenAndServe(":8084", nil)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down notification-service")
}
