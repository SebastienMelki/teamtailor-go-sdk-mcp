.PHONY: all generate build lint lint-fix buf-lint test clean release install-tools deps check apitest mcp-build

# Default target
all: generate build

# Install required tools
install-tools:
	go install github.com/SebastienMelki/sebuf/cmd/protoc-gen-go-client@latest
	go install github.com/SebastienMelki/sebuf/cmd/protoc-gen-openapiv3@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Update buf dependencies
deps:
	buf dep update

# Generate Go code and OpenAPI specs from proto files
generate:
	buf generate
	@echo "Running goimports on generated files..."
	@goimports -w internal/gen/

# Build the Go code
build:
	go build ./...

# Run Go linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Run Go linter with auto-fix
lint-fix:
	@echo "Running linter with auto-fix..."
	@golangci-lint run --fix ./...

# Run buf lint on proto files
buf-lint:
	buf lint

# Run tests
test:
	go test ./...

# Run the apitest integration harness against the live Teamtailor API.
# Requires TEAMTAILOR_API_KEY (and optionally TEAMTAILOR_REGION, TEAMTAILOR_API_VERSION) in .env or env.
apitest:
	go run ./cmd/apitest

# Build the MCP stdio binary
mcp-build:
	go build -o ./bin/teamtailor-mcp ./mcp/cmd/teamtailor-mcp

# Clean generated files
clean:
	rm -rf internal/gen/ docs/ bin/

# Release - creates a new git tag and pushes it
# Usage: make release VERSION=v1.0.0
release:
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=v1.0.0)
endif
	@echo "Creating release $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "Release $(VERSION) created and pushed."

# Run all checks (buf-lint, lint, generate, build, test)
check: buf-lint lint generate build test
