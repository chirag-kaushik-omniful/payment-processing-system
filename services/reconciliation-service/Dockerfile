FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/reconciliation-service/ services/reconciliation-service/
WORKDIR /app/services/reconciliation-service
RUN go mod download && CGO_ENABLED=0 go build -o /reconciliation-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /reconciliation-service /reconciliation-service
EXPOSE 8009
CMD ["/reconciliation-service"]
