FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/webhook-service/ services/webhook-service/
WORKDIR /app/services/webhook-service
RUN go mod download && CGO_ENABLED=0 go build -o /webhook-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /webhook-service /webhook-service
EXPOSE 8008
CMD ["/webhook-service"]
