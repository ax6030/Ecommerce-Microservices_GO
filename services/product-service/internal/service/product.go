package service

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/jason/ecommerce/product-service/internal/cache"
	"github.com/jason/ecommerce/product-service/internal/model"
	"github.com/jason/ecommerce/product-service/internal/repository"
)

type ProductService struct {
	repo   *repository.ProductRepository
	cache  *cache.ProductCache
	logger *slog.Logger
}

func NewProductService(repo *repository.ProductRepository, c *cache.ProductCache, logger *slog.Logger) *ProductService {
	return &ProductService{repo: repo, cache: c, logger: logger}
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	if p, err := s.cache.GetProduct(ctx, id); err == nil {
		return p, nil
	}
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cacheErr := s.cache.SetProduct(ctx, p); cacheErr != nil {
		s.logger.Warn("failed to cache product", "id", id, "error", cacheErr)
	}
	return p, nil
}

func (s *ProductService) ListProducts(ctx context.Context, page, limit int) ([]model.Product, int, error) {
	if products, total, err := s.cache.GetList(ctx, page, limit); err == nil {
		return products, total, nil
	}
	products, total, err := s.repo.List(ctx, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	if cacheErr := s.cache.SetList(ctx, page, limit, products, total); cacheErr != nil {
		s.logger.Warn("failed to cache product list", "error", cacheErr)
	}
	return products, total, nil
}

func (s *ProductService) DeductStock(ctx context.Context, id string, qty int32) (int32, error) {
	remaining, err := s.repo.DeductStock(ctx, id, qty)
	if err != nil {
		return 0, err
	}
	// invalidate cache after stock change
	if cacheErr := s.cache.DeleteProduct(ctx, id); cacheErr != nil && cacheErr != redis.Nil {
		s.logger.Warn("failed to invalidate product cache", "id", id, "error", cacheErr)
	}
	return remaining, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, p *model.Product) error {
	return s.repo.Create(ctx, p)
}
