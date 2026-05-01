package handler

import (
	"context"

	pb "github.com/jason/ecommerce/proto/order"
	"github.com/jason/ecommerce/order-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCHandler struct {
	pb.UnimplementedOrderServiceServer
	svc *service.OrderService
}

func NewOrderGRPCHandler(svc *service.OrderService) *OrderGRPCHandler {
	return &OrderGRPCHandler{svc: svc}
}

func (h *OrderGRPCHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	var items []service.ItemInput
	for _, item := range req.Items {
		items = append(items, service.ItemInput{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.svc.CreateOrder(ctx, service.CreateOrderInput{
		UserID: req.UserId,
		Items:  items,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return &pb.CreateOrderResponse{
		Id:          order.ID,
		UserId:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	}, nil
}

func (h *OrderGRPCHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	order, err := h.svc.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	resp := &pb.GetOrderResponse{
		Id:          order.ID,
		UserId:      order.UserID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt.String(),
	}
	for _, item := range order.Items {
		resp.Items = append(resp.Items, &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	return resp, nil
}

func (h *OrderGRPCHandler) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, total, err := h.svc.ListOrders(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list orders: %v", err)
	}

	resp := &pb.ListOrdersResponse{Total: int32(total)}
	for _, order := range orders {
		o := &pb.GetOrderResponse{
			Id:          order.ID,
			UserId:      order.UserID,
			Status:      order.Status,
			TotalAmount: order.TotalAmount,
			CreatedAt:   order.CreatedAt.String(),
		}
		resp.Orders = append(resp.Orders, o)
	}
	return resp, nil
}
