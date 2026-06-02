FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/analytics-service/ services/analytics-service/
WORKDIR /app/services/analytics-service
RUN go mod download && CGO_ENABLED=0 go build -o /analytics-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /analytics-service /analytics-service
EXPOSE 8010
CMD ["/analytics-service"]
