FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY services/product-service/go.mod services/product-service/go.sum ./services/product-service/
RUN cd services/product-service && GOWORK=off go mod download
COPY services/product-service/ ./services/product-service/
RUN cd services/product-service && GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/services/product-service/db/migrations ./db/migrations
CMD ["./server"]
