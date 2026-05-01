package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/jason/ecommerce/product-service/internal/model"
)

const (
	productTTL = 5 * time.Minute
	listTTL    = 1 * time.Minute
)

type ProductCache struct {
	rdb *redis.Client
}

func NewProductCache(addr string) *ProductCache {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &ProductCache{rdb: rdb}
}

func (c *ProductCache) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	val, err := c.rdb.Get(ctx, productKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var p model.Product
	if err := json.Unmarshal(val, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *ProductCache) SetProduct(ctx context.Context, p *model.Product) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, productKey(p.ID), b, productTTL).Err()
}

func (c *ProductCache) DeleteProduct(ctx context.Context, id string) error {
	return c.rdb.Del(ctx, productKey(id)).Err()
}

func (c *ProductCache) GetList(ctx context.Context, page, limit int) ([]model.Product, int, error) {
	key := listKey(page, limit)
	data, err := c.rdb.HMGet(ctx, key, "products", "total").Result()
	if err != nil || data[0] == nil || data[1] == nil {
		return nil, 0, redis.Nil
	}
	var products []model.Product
	if err := json.Unmarshal([]byte(data[0].(string)), &products); err != nil {
		return nil, 0, err
	}
	var total int
	if err := json.Unmarshal([]byte(data[1].(string)), &total); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (c *ProductCache) SetList(ctx context.Context, page, limit int, products []model.Product, total int) error {
	key := listKey(page, limit)
	pb, err := json.Marshal(products)
	if err != nil {
		return err
	}
	tb, err := json.Marshal(total)
	if err != nil {
		return err
	}
	pipe := c.rdb.Pipeline()
	pipe.HSet(ctx, key, "products", pb, "total", tb)
	pipe.Expire(ctx, key, listTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func productKey(id string) string {
	return fmt.Sprintf("product:%s", id)
}

func listKey(page, limit int) string {
	return fmt.Sprintf("products:list:%d:%d", page, limit)
}
