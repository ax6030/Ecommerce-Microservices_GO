FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY services/order-service/ ./services/order-service/
COPY go.work go.work.sum* ./
RUN cd services/order-service && CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/services/order-service/db/migrations ./db/migrations
EXPOSE 50053 8083
CMD ["./server"]
