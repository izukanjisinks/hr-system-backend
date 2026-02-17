APP_NAME=hr-system
MAIN=./cmd/api/main.go
MIGRATE_DIR=./migrations
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: build run dev tidy migrate-up migrate-down

build:
	go build -o bin/$(APP_NAME) $(MAIN)

run: build
	./bin/$(APP_NAME)

dev:
	go run $(MAIN)

tidy:
	go mod tidy

test:
	go test ./...

migrate-up:
	@for f in $(MIGRATE_DIR)/*.up.sql; do \
		echo "Applying $$f"; \
		psql "$(DB_URL)" -f "$$f"; \
	done

migrate-down:
	@for f in $$(ls -r $(MIGRATE_DIR)/*.down.sql); do \
		echo "Rolling back $$f"; \
		psql "$(DB_URL)" -f "$$f"; \
	done
