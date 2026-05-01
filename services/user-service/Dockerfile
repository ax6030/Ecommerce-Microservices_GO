FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
RUN cd services/user-service && GOWORK=off go mod download
COPY services/user-service/ ./services/user-service/
RUN cd services/user-service && GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/services/user-service/db/migrations ./db/migrations
CMD ["./server"]
