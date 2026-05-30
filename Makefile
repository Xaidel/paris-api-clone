APP_BIN=bin/server
MIGRATE_BIN=bin/migrate

.PHONY: tidy fmt test vet lint build run migrate-up migrate-down migrate-version docker-up docker-down

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

vet:
	go vet ./...

lint: fmt vet test

build:
	mkdir -p bin
	go build -o $(APP_BIN) ./cmd/server
	go build -o $(MIGRATE_BIN) ./cmd/migrate

run:
	ENV=development go run ./cmd/server

migrate-up:
	ENV=development go run ./cmd/migrate up

migrate-down:
	ENV=development go run ./cmd/migrate down

migrate-version:
	ENV=development go run ./cmd/migrate version

docker-up:
	docker compose up --build

docker-down:
	docker compose down --remove-orphans
