FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/auth-service/ services/auth-service/
WORKDIR /app/services/auth-service
RUN go mod download && CGO_ENABLED=0 go build -o /auth-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /auth-service /auth-service
EXPOSE 8001
CMD ["/auth-service"]
