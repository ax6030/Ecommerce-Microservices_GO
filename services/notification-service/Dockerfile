FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY services/notification-service/ ./services/notification-service/
COPY go.work go.work.sum* ./
RUN cd services/notification-service && CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
EXPOSE 8084
CMD ["./server"]
