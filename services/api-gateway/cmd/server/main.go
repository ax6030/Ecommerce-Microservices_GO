package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	pb_order "github.com/jason/ecommerce/proto/order"
	pb_product "github.com/jason/ecommerce/proto/product"
	pb_user "github.com/jason/ecommerce/proto/user"
	"github.com/jason/ecommerce/api-gateway/internal/handler"
	"github.com/jason/ecommerce/api-gateway/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "api-gateway")

	userAddr := getEnv("USER_SERVICE_ADDR", "localhost:50051")
	productAddr := getEnv("PRODUCT_SERVICE_ADDR", "localhost:50052")
	orderAddr := getEnv("ORDER_SERVICE_ADDR", "localhost:50053")

	userConn := mustDial(userAddr, logger)
	defer userConn.Close()
	productConn := mustDial(productAddr, logger)
	defer productConn.Close()
	orderConn := mustDial(orderAddr, logger)
	defer orderConn.Close()

	userClient := pb_user.NewUserServiceClient(userConn)
	productClient := pb_product.NewProductServiceClient(productConn)
	orderClient := pb_order.NewOrderServiceClient(orderConn)

	authMW := middleware.NewAuthMiddleware(userClient)
	userH := handler.NewUserHandler(userClient)
	productH := handler.NewProductHandler(productClient)
	orderH := handler.NewOrderHandler(orderClient)

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "api-gateway"})
	})

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", userH.Register)
		auth.POST("/login", userH.Login)
	}

	users := r.Group("/api/v1/users", authMW.Authenticate())
	{
		users.GET("/me", userH.GetMe)
	}

	products := r.Group("/api/v1/products")
	{
		products.GET("", productH.ListProducts)
		products.GET("/:id", productH.GetProduct)
		products.POST("", authMW.Authenticate(), productH.CreateProduct)
	}

	orders := r.Group("/api/v1/orders", authMW.Authenticate())
	{
		orders.POST("", orderH.CreateOrder)
		orders.GET("", orderH.ListOrders)
		orders.GET("/:id", orderH.GetOrder)
	}

	port := getEnv("HTTP_PORT", "8080")
	logger.Info("API Gateway starting", "port", port)
	if err := r.Run(":" + port); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func mustDial(addr string, logger *slog.Logger) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect", "addr", addr, "error", err)
		os.Exit(1)
	}
	return conn
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
