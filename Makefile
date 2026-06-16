.PHONY: build run test dev clean docker-up docker-down lint swagger swagger-clean coverage coverage-html

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
