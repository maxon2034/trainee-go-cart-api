BUILD_DIR = cmd
APP_NAME = cart_api

DB_USER = cart_user
DB_ADDR = localhost
DB_PORT = 5432
DB_NAME = cart_db
# DB_PASSWORD must be environmental variable

POSTGRES_DSN = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_ADDR):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: run build test test-integration lint generate migration-up migration-down

generate:
	go generate ./...

migration-up:
	goose -dir ./migrations postgres "$(POSTGRES_DSN)" up

migration-down:
	goose -dir ./migrations postgres "$(POSTGRES_DSN)" down

build:
	@echo "building app"
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd

run: build
	@echo "running $(APP_NAME)"
	./$(BUILD_DIR)/$(APP_NAME)

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

lint:
	golangci-lint run


