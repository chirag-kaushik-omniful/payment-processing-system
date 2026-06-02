FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/api-gateway/ services/api-gateway/
WORKDIR /app/services/api-gateway
RUN go mod download && CGO_ENABLED=0 go build -o /api-gateway ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /api-gateway /api-gateway
EXPOSE 8000
CMD ["/api-gateway"]
