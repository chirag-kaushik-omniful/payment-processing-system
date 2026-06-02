FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/wallet-service/ services/wallet-service/
WORKDIR /app/services/wallet-service
RUN go mod download && CGO_ENABLED=0 go build -o /wallet-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /wallet-service /wallet-service
EXPOSE 8004
CMD ["/wallet-service"]
