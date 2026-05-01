FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY services/notification-service/go.mod services/notification-service/go.sum ./services/notification-service/
RUN cd services/notification-service && GOWORK=off go mod download
COPY services/notification-service/ ./services/notification-service/
RUN cd services/notification-service && GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .

CMD ["./server"]
