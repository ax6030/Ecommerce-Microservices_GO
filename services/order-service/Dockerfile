FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY services/order-service/go.mod services/order-service/go.sum ./services/order-service/
RUN cd services/order-service && GOWORK=off go mod download
COPY services/order-service/ ./services/order-service/
RUN cd services/order-service && GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/services/order-service/db/migrations ./db/migrations
CMD ["./server"]
