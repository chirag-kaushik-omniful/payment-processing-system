.PHONY: all build test lint tidy migrate-up docker-up docker-down web-install web-dev web-build

SERVICES := api-gateway auth-service payment-service provider-service \
	wallet-service ledger-service fraud-service notification-service \
	webhook-service reconciliation-service analytics-service \
	admin-service subscription-service

all: build

build:
	@mkdir -p bin
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		( cd services/$$svc && go build -o ../../bin/$$svc ./cmd/server ) || exit 1; \
	done

test:
	@go work sync
	@go test ./shared/... -count=1
	@for svc in $(SERVICES); do \
		( cd services/$$svc && go test ./... -count=1 ) || exit 1; \
	done

lint:
	@go vet ./shared/... ./services/...

tidy:
	@go work sync
	@cd shared && go mod tidy
	@for svc in $(SERVICES); do cd services/$$svc && go mod tidy && cd ../..; done

migrate-up:
	@./scripts/migrate.sh up

docker-up:
	docker compose -f deployments/docker/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker/docker-compose.yml down

run-gateway:
	cd services/api-gateway && go run ./cmd/server

run-auth:
	cd services/auth-service && go run ./cmd/server

run-payment:
	cd services/payment-service && go run ./cmd/server

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build
