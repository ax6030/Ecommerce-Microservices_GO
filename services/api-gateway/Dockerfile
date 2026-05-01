FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY proto/ ./proto/
COPY services/api-gateway/ ./services/api-gateway/
COPY go.work go.work.sum* ./
RUN cd services/api-gateway && CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
EXPOSE 8080
CMD ["./server"]
