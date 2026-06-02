FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/provider-service/ services/provider-service/
WORKDIR /app/services/provider-service
RUN go mod download && CGO_ENABLED=0 go build -o /provider-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /provider-service /provider-service
EXPOSE 8003
CMD ["/provider-service"]
