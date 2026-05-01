package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	pb "github.com/jason/ecommerce/proto/order"
)

type OrderHandler struct {
	client pb.OrderServiceClient
}

func NewOrderHandler(client pb.OrderServiceClient) *OrderHandler {
	return &OrderHandler{client: client}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Items []struct {
			ProductID string `json:"product_id" binding:"required"`
			Quantity  int32  `json:"quantity" binding:"required,gt=0"`
		} `json:"items" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pbItems []*pb.OrderItem
	for _, item := range req.Items {
		pbItems = append(pbItems, &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	resp, err := h.client.CreateOrder(context.Background(), &pb.CreateOrderRequest{
		UserId: userID,
		Items:  pbItems,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.client.GetOrder(context.Background(), &pb.GetOrderRequest{Id: id})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resp, err := h.client.ListOrders(context.Background(), &pb.ListOrdersRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
