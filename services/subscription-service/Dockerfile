FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/subscription-service/ services/subscription-service/
WORKDIR /app/services/subscription-service
RUN go mod download && CGO_ENABLED=0 go build -o /subscription-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /subscription-service /subscription-service
EXPOSE 8012
CMD ["/subscription-service"]
