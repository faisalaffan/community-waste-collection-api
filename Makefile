.PHONY: build run test dev clean docker-up docker-down lint

build:
	go build -o bin/server ./cmd/server/

run: build
	./bin/server

dev:
	go run ./cmd/server/

test:
	go test ./... -v

clean:
	rm -rf bin/

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

lint:
	go vet ./...
