.PHONY: build run test vet clean clean-data clean-build docker-build docker-up docker-down docker-prod-build docker-prod-up docker-prod-down help

# Default target
help:
	@echo "GoNote - Available Commands:"
	@echo ""
	@echo "  Go Development:"
	@echo "    make build        - Build Go server"
	@echo "    make run          - Run Go server (development)"
	@echo "    make test         - Run Go tests with race detector"
	@echo "    make vet          - Run go vet"
	@echo ""
	@echo "  Data Cleanup:"
	@echo "    make clean-data       - Clean temporary and cache files"
	@echo "    make clean-build      - Clean build artifacts"
	@echo "    make clean            - Full cleanup (data + build)"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-build      - Build Docker image (Go)"
	@echo "    make docker-up         - Start Docker Compose (Go development)"
	@echo "    make docker-down       - Stop Docker Compose"
	@echo "    make docker-prod-build - Build production Docker image"
	@echo "    make docker-prod-up    - Start production Docker (pre-built image)"
	@echo "    make docker-prod-down  - Stop production Docker"
	@echo ""
	@echo "  Tests:"
	@echo "    make test-e2e        - Run Playwright E2E tests"
	@echo "    make test-e2e-ui     - Run tests with Playwright UI"
	@echo ""
	@echo "  Frontend:"
	@echo "    make deps         - Install frontend dependencies"
	@echo "    make css-build    - Build Tailwind CSS"
	@echo "    make css-watch    - Watch and rebuild Tailwind CSS"
	@echo ""
	@echo "  Cleanup:"
	@echo "    make clean        - Clean build artifacts"

# Data cleanup
data-clean:
	rm -rf data/temp/*
	find data/cache -type f -mtime +7 -delete 2>/dev/null || true
	@echo "Cleaned temporary files and old cache"

clean-build:
	rm -rf go/server go/main go/gonote go/*.test go/test.log
	@echo "Cleaned build artifacts"

clean: clean-data clean-build
	@echo "Full cleanup completed"

# Build
build:
	cd go && go build -o ../gonote ./cmd/server

run:
	go run go/cmd/server/main.go --config go/config.yaml

docker-build:
	docker build -f docker/go/Dockerfile -t gonote-go .

docker-up:
	docker-compose -f docker/compose/development.yml up -d

docker-down:
	docker-compose -f docker/compose/development.yml down

docker-prod-build:
	docker-compose -f docker-compose.ghcr.yml build

docker-prod-up:
	docker-compose -f docker-compose.ghcr.yml up -d

docker-prod-down:
	docker-compose -f docker-compose.ghcr.yml down

# Tests
test:
	cd go && STORAGE_NOTES_DIR=../data go test ./...
test-e2e:
	cd tests && npx playwright test
test-e2e-ui:
	cd tests && npx playwright test --ui