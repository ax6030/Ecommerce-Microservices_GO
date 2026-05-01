# Ecommerce Microservices GO

以 Go 實作的電商微服務系統，涵蓋 gRPC 服務間通訊、Redis 快取、RabbitMQ 訊息佇列、JWT 認證等核心技術。

---

## 系統架構

```
Client
  │
  ▼
┌─────────────────┐
│   api-gateway   │  HTTP :8080（Gin）
│  JWT 驗證中介層  │
└────────┬────────┘
         │  gRPC
    ┌────┼────────────┐
    ▼    ▼            ▼
┌───────┐ ┌─────────┐ ┌───────────┐
│ user  │ │ product │ │  order    │
│service│ │ service │ │  service  │
│:50051 │ │ :50052  │ │  :50053   │
└───┬───┘ └────┬────┘ └─────┬─────┘
    │          │             │
    ▼          ▼             ▼
 user-db  product-db      order-db
(MariaDB) (MariaDB)      (MariaDB)
               │             │
               ▼             ▼
            Redis        RabbitMQ
          (快取層)       (訊息佇列)
                             │
                             ▼
                    ┌─────────────────┐
                    │ notification    │
                    │ service         │
                    │（訂閱 + 模擬寄信）│
                    └─────────────────┘
```

### 服務一覽

| 服務 | 職責 | Port |
|------|------|------|
| api-gateway | HTTP 入口、JWT 驗證、路由轉發至各 gRPC 服務 | HTTP 8080 |
| user-service | 使用者註冊、登入、JWT 發行 | gRPC 50051, Health 8081 |
| product-service | 商品 CRUD、庫存扣減、Redis 快取 | gRPC 50052, Health 8082 |
| order-service | 建立訂單、呼叫 product-service 扣庫存、發布事件 | gRPC 50053, Health 8083 |
| notification-service | 消費 RabbitMQ 訂單事件、模擬 Email 通知 | — |

### 基礎設施

| 元件 | 用途 | Port |
|------|------|------|
| MariaDB × 3 | 各服務獨立資料庫（user/product/order） | 內部 |
| Redis 7 | product-service 商品快取 | 6379 |
| RabbitMQ 3.13 | 訂單事件佇列（含 DLQ）| 5672, 管理介面 15672 |

---

## 快速開始

### 前置需求

- Docker & Docker Compose
- Go 1.22+（本地開發）

### 啟動所有服務

```bash
# 複製專案
git clone <repo-url>
cd Ecommerce-Microservices_GO

# 啟動（首次會建置所有 Docker image）
make up

# 查看 logs
make logs

# 停止
make down

# 停止並刪除 volumes（清空所有資料）
make clean
```

服務啟動後：
- API Gateway：http://localhost:8080
- RabbitMQ 管理介面：http://localhost:15672（帳號 `guest` / 密碼 `guest`）

---

## API 文件

所有需要認證的 API 須在 Header 帶入：
```
Authorization: Bearer <jwt_token>
```

### 使用者

| Method | Path | 說明 | 認證 |
|--------|------|------|------|
| POST | `/api/v1/auth/register` | 註冊新帳號 | 否 |
| POST | `/api/v1/auth/login` | 登入，回傳 JWT | 否 |
| GET | `/api/v1/users/me` | 取得目前登入使用者資訊 | 是 |

**POST /api/v1/auth/register**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "Jason"
}
```

**POST /api/v1/auth/login**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```
回應：
```json
{
  "token": "eyJhbGci..."
}
```

### 商品

| Method | Path | 說明 | 認證 |
|--------|------|------|------|
| GET | `/api/v1/products` | 商品列表（分頁）| 否 |
| GET | `/api/v1/products/:id` | 取得單一商品 | 否 |
| POST | `/api/v1/products` | 新增商品 | 是 |

**GET /api/v1/products?page=1&limit=20**
```json
{
  "products": [...],
  "total": 100
}
```

**POST /api/v1/products**
```json
{
  "name": "iPhone 16",
  "description": "最新款 iPhone",
  "price": 35900.00,
  "stock": 50
}
```

### 訂單

| Method | Path | 說明 | 認證 |
|--------|------|------|------|
| POST | `/api/v1/orders` | 建立訂單 | 是 |
| GET | `/api/v1/orders/:id` | 取得訂單詳情 | 是 |
| GET | `/api/v1/orders` | 取得我的訂單列表 | 是 |

**POST /api/v1/orders**
```json
{
  "items": [
    { "product_id": "uuid-xxx", "quantity": 2 },
    { "product_id": "uuid-yyy", "quantity": 1 }
  ]
}
```

建立訂單流程：
1. order-service 向 product-service 確認庫存並扣減
2. 訂單寫入 order-db
3. 發布 `order.created` 事件至 RabbitMQ
4. notification-service 消費事件，記錄通知 log（模擬寄信）

---

## 技術設計

### Redis 快取（product-service）

採用 cache-through 策略：

```
讀取商品：
  Cache hit → 直接回傳
  Cache miss → 查 DB → 寫入 Cache（TTL 5min）→ 回傳

讀取列表：
  Cache hit → 直接回傳
  Cache miss → 查 DB → 寫入 Cache（TTL 1min）→ 回傳

扣庫存（DeductStock）：
  DB 寫入（行鎖 FOR UPDATE）→ 刪除 Cache key（失效）
```

### RabbitMQ 拓樸

```
order-service
    │ Publish: routing key = "order.created"
    ▼
Exchange: order.events (direct, durable)
    │
    ▼
Queue: notification.order.created (durable)
    │ 訊息處理失敗（nack）
    ▼
DLX Exchange: order.events.dlx
    │
    ▼
DLQ: notification.order.created.dlq
```

消費端採手動 Ack，確保訊息不遺失；失敗訊息路由至 DLQ 供後續排查。

### gRPC Proto

```
proto/
├── user/user.proto     Register, Login, GetUser
├── product/product.proto  GetProduct, ListProducts, CreateProduct, DeductStock
└── order/order.proto   CreateOrder, GetOrder, ListOrders
```

重新產生 proto（需安裝 protoc）：
```bash
make proto
```

---

## 本地開發

```bash
# 使用 go.work（monorepo 模式）
go work sync

# 只跑基礎設施（DB + Redis + RabbitMQ）
docker compose up -d user-db product-db order-db redis rabbitmq

# 個別啟動服務（以 product-service 為例）
cd services/product-service
PRODUCT_DB_DSN="user:password@tcp(localhost:3306)/productdb?parseTime=true" \
REDIS_ADDR="localhost:6379" \
go run cmd/server/main.go

# 執行測試
make test
```

---

## 健康檢查

各服務均提供 `/healthz` HTTP 端點：

```bash
curl http://localhost:8081/healthz  # user-service
curl http://localhost:8082/healthz  # product-service
curl http://localhost:8083/healthz  # order-service
curl http://localhost:8080/healthz  # api-gateway
```

---

## 專案結構

```
Ecommerce-Microservices_GO/
├── docker-compose.yml
├── go.work
├── Makefile
├── PLAN.md
├── proto/
│   ├── user/
│   ├── product/
│   └── order/
└── services/
    ├── api-gateway/
    ├── user-service/
    ├── product-service/
    │   └── internal/
    │       ├── cache/       # Redis 快取層
    │       ├── service/     # 業務邏輯 + 快取協調
    │       ├── repository/  # DB 存取
    │       ├── handler/     # gRPC handler
    │       └── model/
    ├── order-service/
    │   └── internal/
    │       └── messaging/   # RabbitMQ publisher
    └── notification-service/
        └── internal/
            └── messaging/   # RabbitMQ consumer + DLQ
```
