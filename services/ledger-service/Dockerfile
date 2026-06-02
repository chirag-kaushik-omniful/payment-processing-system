FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY shared/ shared/
COPY services/ledger-service/ services/ledger-service/
WORKDIR /app/services/ledger-service
RUN go mod download && CGO_ENABLED=0 go build -o /ledger-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /ledger-service /ledger-service
EXPOSE 8005
CMD ["/ledger-service"]
