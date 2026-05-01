FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY services/user-service/ ./services/user-service/
COPY go.work go.work.sum* ./
RUN cd services/user-service && CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/services/user-service/db/migrations ./db/migrations
EXPOSE 50051 8081
CMD ["./server"]
