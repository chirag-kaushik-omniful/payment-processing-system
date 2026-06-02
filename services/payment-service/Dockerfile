FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/payment-service/ services/payment-service/
WORKDIR /app/services/payment-service
RUN go mod download && CGO_ENABLED=0 go build -o /payment-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /payment-service /payment-service
EXPOSE 8002
CMD ["/payment-service"]
