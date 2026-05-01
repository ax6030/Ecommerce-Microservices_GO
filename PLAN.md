# Ecommerce Microservices GO — 開發計畫

## 專案概述

以 Go 實作的電商微服務系統，作為學習 Go 微服務架構的實踐專案。
涵蓋 gRPC 通訊、Redis 快取、RabbitMQ 訊息佇列、JWT 認證等核心技術。

---

## 分支規劃

| 分支 | 用途 |
|------|------|
| `main` | 穩定發布版本 |
| `develop` | 開發主分支 |
| `feature/xxx` | 功能開發分支 |

---

## Phase 4：微服務基礎架構

### 任務清單

| # | 任務 | 分支 | 狀態 |
|---|------|------|------|
| 4-1 | 建立 monorepo 結構（go.work + proto）| `feature/phase4-init` | ✅ 完成 |
| 4-2 | user-service（註冊/登入/JWT）| `feature/phase4-init` | ✅ 完成 |
| 4-3 | product-service（CRUD + 扣庫存）| `feature/phase4-init` | ✅ 完成 |
| 4-4 | order-service（建單 + 呼叫 product gRPC）| `feature/phase4-init` | ✅ 完成 |
| 4-5 | notification-service（訂閱訊息 + 模擬寄信）| `feature/phase4-init` | ✅ 完成 |
| 4-6 | api-gateway（Gin HTTP → gRPC proxy）| `feature/phase4-init` | ✅ 完成 |
| 4-7 | Docker Compose 完整部署（含 DB migration）| `feature/docker-build-fixes` | ✅ 完成 |
| 4-8 | Redis 快取層（product-service cache-through）| `feature/redis-product-cache` | ✅ 完成 |
| 4-9 | RabbitMQ 訊息佇列（取代 NATS，含 DLQ）| `feature/redis-product-cache` | ✅ 完成 |

### 技術細節

**Redis 快取策略（product-service）**
- `GetProduct`：cache-through，TTL 5 分鐘
- `ListProducts`：cache-through（以 page+limit 為 key），TTL 1 分鐘
- `DeductStock`：DB 寫入後刪除 cache key（Cache Invalidation）

**RabbitMQ 拓樸**
- Exchange：`order.events`（direct, durable）
- Queue：`notification.order.created`（durable, 綁定 `order.created`）
- DLX：`order.events.dlx` + DLQ `notification.order.created.dlq`
- 消費端手動 Ack/Nack，失敗訊息路由至 DLQ

---

## Phase 5：可觀測性與測試（待開發）

| # | 任務 | 分支 | 狀態 |
|---|------|------|------|
| 5-1 | Prometheus metrics（各服務暴露 /metrics）| `feature/prometheus-metrics` | 🔲 待開始 |
| 5-2 | Grafana Dashboard（容器化）| `feature/prometheus-metrics` | 🔲 待開始 |
| 5-3 | /healthz 整合測試（Testcontainers）| `feature/integration-tests` | 🔲 待開始 |
| 5-4 | Service 層 Unit Test（含 mock）| `feature/unit-tests` | 🔲 待開始 |

---

## Phase 6：進階功能（規劃中）

| # | 任務 | 說明 |
|---|------|------|
| 6-1 | CI/CD Pipeline | GitLab CI + Jenkinsfile rolling update |
| 6-2 | Rate Limiting | api-gateway 加入限流中介層 |
| 6-3 | Distributed Tracing | OpenTelemetry + Jaeger |
| 6-4 | gRPC 錯誤標準化 | 統一 status code 與 error detail |

---

## Git Commit 記錄（develop 分支）

```
4162c2a feat: merge feature/redis-product-cache → develop
f1dd7cb feat(backend): 以 RabbitMQ 取代 NATS，實作訂單通知持久化佇列與 DLQ
eab3e58 feat(backend): 新增 Redis 快取層至 product-service
3e061d6 fix(infra): merge feature/docker-build-fixes → develop
316f0b3 fix(infra): 修復 Docker 建置與 migration 問題
557e093 feat(backend): 建立 Phase 4 微服務專案結構與所有服務實作
```
