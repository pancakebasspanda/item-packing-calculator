.PHONY: up down build logs restart clean test test-integration test-fuzz test-all

# Start both frontend and backend in the background
up:
	docker-compose up -d

# Stop both containers
down:
	docker-compose down

# Rebuild both images and start them (use this after changing code)
build:
	docker-compose up --build -d

# View live logs from both containers
logs:
	docker-compose logs -f

# Restart the services
restart:
	docker-compose restart

# Completely wipe containers, networks, and images to start fresh
clean:
	docker-compose down -v --rmi all

# Runs only standard unit tests (ignores files with //go:build integration)
test:
	@echo "Running unit tests..."
	go test -race -v ./...

# Runs only the integration tests inside the integration_tests folder
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./integration_tests/...

# Runs only the fuzz inside the fuzz_tests folder in the core service
test-fuzz:
	@echo "Running fuzz tests..."
	go test -v -fuzz=FuzzCalculate -fuzztime=10s ./internal/core/service/fuzz_test

# Runs both unit tests and integration tests together
test-all:
	@echo "Running all tests (Unit + Integration)..."
	go test -v -tags=integration ./...