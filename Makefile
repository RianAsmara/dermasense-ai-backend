.PHONY: dev run build docker-up docker-down docker-logs deps clean test fmt lint

# Development - run backend with hot reload (requires air)
dev:
	go run cmd/server/main.go

# Run the compiled backend binary
run:
	./server

# Build the backend binary
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o server ./cmd/server/

# Docker Compose (infrastructure only: postgres + redis)
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# Download Go dependencies
deps:
	go mod tidy && go mod download

# Clean build artifacts
clean:
	rm -f server
	docker compose down -v --rmi local

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	go vet ./...
