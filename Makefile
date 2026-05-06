.PHONY: run build test migrate migrate-down migrate-status lint docker health

API_DIR := bpcl-portal-api

run:
	cd $(API_DIR) && go run ./cmd/api

build:
	cd $(API_DIR) && go build -o ../bin/api ./cmd/api

test:
	cd $(API_DIR) && go test ./... -v -race

migrate:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down 1

migrate-status:
	./scripts/migrate.sh version

lint:
	cd $(API_DIR) && golangci-lint run ./...

docker:
	docker-compose up -d

health:
	curl -s http://localhost:8080/health | jq .
