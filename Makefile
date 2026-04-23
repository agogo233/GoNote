.PHONY: build run test vet clean docker-build docker-up docker-down deps help

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

# ==================== Go Development ====================

build:
	cd go && go build -o server ./cmd/server

run:
	cd go && go run cmd/server/main.go

test:
	cd go && go test ./... -race

vet:
	cd go && go vet ./...

# ==================== Docker ====================

docker-build:
	docker build -f docker/go/Dockerfile -t gonote:dev .

docker-up:
	docker-compose -f docker/compose/development.yml up

docker-down:
	docker-compose -f docker/compose/development.yml down

docker-prod-build:
	docker build -f docker/go/Dockerfile -t gonote:latest .

docker-prod-up:
	docker-compose -f docker/compose/production.yml up -d

docker-prod-down:
	docker-compose -f docker/compose/production.yml down

# ==================== Frontend ====================

deps:
	npm install

css-build:
	npx tailwindcss -i ./build/tailwind/input.css -o ./shared/frontend/libs/tailwind/tailwind.css --minify

css-watch:
	npx tailwindcss -i ./build/tailwind/input.css -o ./shared/frontend/libs/tailwind/tailwind.css --watch

# ==================== Tests ====================

test-e2e:
	npx playwright test --config=tests/playwright.config.ts

test-e2e-ui:
	npx playwright test --config=tests/playwright.config.ts --ui

# ==================== Cleanup ====================

clean:
	rm -rf go/server go/main go/gonote
	go clean -cache
	rm -rf node_modules
