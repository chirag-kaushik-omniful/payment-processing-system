FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/notification-service/ services/notification-service/
WORKDIR /app/services/notification-service
RUN go mod download && CGO_ENABLED=0 go build -o /notification-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /notification-service /notification-service
EXPOSE 8007
CMD ["/notification-service"]
