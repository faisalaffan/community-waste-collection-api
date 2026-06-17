.PHONY: build run test dev clean docker-up docker-down lint swagger swagger-clean coverage coverage-html db-url schema-apply schema-diff schema-inspect schema-dump k8s-build k8s-diff k8s-apply k8s-delete

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

PKGS := $(shell go list ./... | grep -v /docs | grep -v /cmd/server)

coverage:
	go test $(PKGS) -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out | grep -v '_test.go'

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

schema-dump:
	atlas schema inspect --config $(ATLAS_CONFIG) --env local --format '{{ sql . }}' > migrations/schema.sql
	@echo "DDL dumped to migrations/schema.sql"

# ── Kubernetes ──
K8S_OVERLAY := production
K8S_CTX := vps-malang
K8S_SECRET_ENV := .env.secret

$(K8S_SECRET_ENV):
	@echo "DB_USER=$(DB_USER)" > $@
	@echo "DB_PASSWORD=$(DB_PASSWORD)" >> $@
	@echo "S3_ACCESS_KEY=$(S3_ACCESS_KEY)" >> $@
	@echo "S3_SECRET_KEY=$(S3_SECRET_KEY)" >> $@

k8s-build: $(K8S_SECRET_ENV)
	kustomize build k8s/overlays/$(K8S_OVERLAY)

k8s-diff: $(K8S_SECRET_ENV)
	kustomize build k8s/overlays/$(K8S_OVERLAY) | kubectl --context=$(K8S_CTX) diff -f -

k8s-apply: $(K8S_SECRET_ENV)
	kustomize build k8s/overlays/$(K8S_OVERLAY) | kubectl --context=$(K8S_CTX) apply -f -

k8s-delete: $(K8S_SECRET_ENV)
	kustomize build k8s/overlays/$(K8S_OVERLAY) | kubectl --context=$(K8S_CTX) delete -f -
