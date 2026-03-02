.PHONY: dev build check lint cross release install-tools

# Dev: air (Go hot reload) + Vite in parallel
dev:
	@echo "Starting development servers..."
	@cd client && pnpm install --frozen-lockfile 2>/dev/null || pnpm install
	@trap 'kill 0' INT; \
	  (air) & \
	  (cd client && pnpm dev) & \
	  wait

# Build: vite build → server/public/; then go build
build:
	@echo "Building client..."
	cd client && pnpm install && pnpm build
	@echo "Building server..."
	go build -o dist/vibecodepc ./server/main.go

# Check: go vet + vue-tsc
check:
	go vet ./...
	cd client && pnpm exec vue-tsc --noEmit

# Lint: golangci-lint + eslint
lint:
	golangci-lint run ./...
	cd client && pnpm exec eslint src/

# Cross-compile for ARM64
cross:
	GOOS=linux GOARCH=arm64 go build -o dist/vibecodepc-arm64 ./server/main.go

# Build release binaries for all architectures (arm64, arm, amd64)
release:
	bash scripts/build-release.sh

# Install dev tools
install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
