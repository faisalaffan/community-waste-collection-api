.PHONY: build run test dev clean docker-up docker-down lint swagger swagger-clean coverage coverage-html db-url schema-apply schema-diff schema-inspect

-include .env
export

DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

build:
	go build -o bin/server ./cmd/server/

run: build
	./bin/server

dev:
	go run ./cmd/server/

test:
	go test ./... -v

test-short:
	go test ./... -v -short

clean:
	rm -rf bin/

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

lint:
	go vet ./...

swagger:
	swag init -g cmd/server/main.go -o docs/

swagger-clean:
	rm -rf docs/docs.go docs/swagger.json docs/swagger.yaml
	swag init -g cmd/server/main.go -o docs/

coverage:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open: open coverage.html"

deps:
	go mod tidy
	go mod download

all: lint test build

# ── Atlas Schema (declarative) ──
ATLAS_CONFIG := file://migrations/atlas.hcl

db-url:
	@echo "$(DATABASE_URL)"

schema-apply:
	atlas schema apply --config $(ATLAS_CONFIG) --env local --to file://migrations/schema.pg.hcl

schema-diff:
	atlas schema apply --config $(ATLAS_CONFIG) --env local --to file://migrations/schema.pg.hcl --dry-run

schema-inspect:
	atlas schema inspect --config $(ATLAS_CONFIG) --env local
