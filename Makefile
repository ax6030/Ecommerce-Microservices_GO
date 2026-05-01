.PHONY: proto build up down logs clean

PROTO_DIR=proto
OUT_DIR=.

proto:
	protoc --go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/user/user.proto \
		$(PROTO_DIR)/product/product.proto \
		$(PROTO_DIR)/order/order.proto

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --remove-orphans

test:
	@for svc in user-service product-service order-service notification-service api-gateway; do \
		echo "Testing $$svc..."; \
		cd services/$$svc && go test ./... && cd ../..; \
	done
