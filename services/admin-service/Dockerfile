FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/admin-service/ services/admin-service/
WORKDIR /app/services/admin-service
RUN go mod download && CGO_ENABLED=0 go build -o /admin-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /admin-service /admin-service
EXPOSE 8011
CMD ["/admin-service"]
