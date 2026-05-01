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
	pb "github.com/jason/ecommerce/proto/product"
	"github.com/jason/ecommerce/product-service/internal/handler"
	"github.com/jason/ecommerce/product-service/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "product-service")

	dsn := os.Getenv("PRODUCT_DB_DSN")
	if dsn == "" {
		dsn = "user:password@tcp(localhost:3306)/productdb?parseTime=true"
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

	repo := repository.NewProductRepository(db)
	grpcHandler := handler.NewProductGRPCHandler(repo)

	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"product-service"}`))
		})
		logger.Info("health server listening", "port", "8082")
		http.ListenAndServe(":8082", nil)
	}()

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	server := grpc.NewServer()
	pb.RegisterProductServiceServer(server, grpcHandler)
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
