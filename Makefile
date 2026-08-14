.PHONY: help build run test frontend docker clean lint fmt vet

BINARY  := booksync
IMAGE   := ghcr.io/jeeftor/booksync
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
	/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: frontend ## Build the booksync binary (embeds the frontend)
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

run: ## Run the server locally (serves API only unless `make frontend` has been run)
	go run . serve

frontend: ## Build the frontend and copy it into internal/webui/dist for embedding
	cd frontend && bun install --frozen-lockfile && bun run build
	rm -rf internal/webui/dist
	mkdir -p internal/webui/dist
	cp -r frontend/dist/. internal/webui/dist/

test: ## Run the test suite
	go test ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format Go code
	gofmt -w .

vet: ## Run go vet
	go vet ./...

docker: ## Build the Docker image
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE) -t $(IMAGE):dev .

clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -rf frontend/dist internal/webui/dist/*
	touch internal/webui/dist/.gitkeep
