FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/fraud-service/ services/fraud-service/
WORKDIR /app/services/fraud-service
RUN go mod download && CGO_ENABLED=0 go build -o /fraud-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /fraud-service /fraud-service
EXPOSE 8006
CMD ["/fraud-service"]
