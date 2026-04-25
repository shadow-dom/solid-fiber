.PHONY: dev dev-frontend dev-backend build build-frontend build-backend run clean install

FRONTEND_DIR := frontend/app
BACKEND_DIR  := backend
BIN          := bin/server

install:
	cd $(FRONTEND_DIR) && bun install
	cd $(BACKEND_DIR) && go mod tidy

# Dev: Vite (HMR) on :5173 proxies /api to Fiber on :3000.
# Open http://localhost:5173 during development.
dev:
	@trap 'kill 0' INT TERM EXIT; \
		( cd $(FRONTEND_DIR) && bun run dev ) & \
		( cd $(BACKEND_DIR)  && go run . ) & \
		wait

# Prod: build SPA into backend/web/dist, then build single Go binary that embeds it.
build: build-frontend build-backend

build-frontend:
	cd $(FRONTEND_DIR) && bun run build

build-backend:
	mkdir -p bin
	cd $(BACKEND_DIR) && go build -o ../$(BIN) .

run: build
	./$(BIN)

clean:
	rm -rf $(BIN) backend/web/dist/*
	touch backend/web/dist/.gitkeep
