package handler

import (
	"context"

	pb "github.com/jason/ecommerce/proto/product"
	"github.com/jason/ecommerce/product-service/internal/model"
	"github.com/jason/ecommerce/product-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductGRPCHandler struct {
	pb.UnimplementedProductServiceServer
	svc *service.ProductService
}

func NewProductGRPCHandler(svc *service.ProductService) *ProductGRPCHandler {
	return &ProductGRPCHandler{svc: svc}
}

func (h *ProductGRPCHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, err := h.svc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found")
	}
	return toProtoProduct(p), nil
}

func (h *ProductGRPCHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}

	products, total, err := h.svc.ListProducts(ctx, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	var pbProducts []*pb.GetProductResponse
	for _, p := range products {
		p := p
		pbProducts = append(pbProducts, toProtoProduct(&p))
	}
	return &pb.ListProductsResponse{Products: pbProducts, Total: int32(total)}, nil
}

func (h *ProductGRPCHandler) DeductStock(ctx context.Context, req *pb.DeductStockRequest) (*pb.DeductStockResponse, error) {
	remaining, err := h.svc.DeductStock(ctx, req.ProductId, req.Quantity)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	return &pb.DeductStockResponse{Success: true, RemainingStock: remaining}, nil
}

func (h *ProductGRPCHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	p := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}
	if err := h.svc.CreateProduct(ctx, p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}
	return &pb.CreateProductResponse{
		Id:    p.ID,
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}, nil
}

func toProtoProduct(p *model.Product) *pb.GetProductResponse {
	return &pb.GetProductResponse{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
	}
}
