package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	pb_order "github.com/jason/ecommerce/proto/order"
	pb_product "github.com/jason/ecommerce/proto/product"
	"github.com/jason/ecommerce/order-service/internal/handler"
	"github.com/jason/ecommerce/order-service/internal/messaging"
	"github.com/jason/ecommerce/order-service/internal/repository"
	"github.com/jason/ecommerce/order-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "order-service")

	dsn := os.Getenv("ORDER_DB_DSN")
	if dsn == "" {
		dsn = "user:password@tcp(localhost:3306)/orderdb?parseTime=true"
	}

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		logger.Error("failed to open db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db.DB); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	publisher, err := messaging.NewPublisher(rabbitmqURL)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()
	logger.Info("connected to RabbitMQ", "url", rabbitmqURL)

	productAddr := os.Getenv("PRODUCT_SERVICE_ADDR")
	if productAddr == "" {
		productAddr = "localhost:50052"
	}
	productConn, err := grpc.NewClient(productAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to product-service", "error", err)
		os.Exit(1)
	}
	defer productConn.Close()

	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo, pb_product.NewProductServiceClient(productConn), publisher)
	grpcHandler := handler.NewOrderGRPCHandler(svc)

	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"order-service"}`))
		})
		logger.Info("health server listening", "port", "8083")
		http.ListenAndServe(":8083", nil)
	}()

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	server := grpc.NewServer()
	pb_order.RegisterOrderServiceServer(server, grpcHandler)
	reflection.Register(server)

	logger.Info("gRPC server listening", "port", grpcPort)
	if err := server.Serve(lis); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func runMigrations(db *sql.DB) error {
	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "mysql", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
