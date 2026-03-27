.PHONY: build clean run test install

# Build the project
build:
	go build -o bin/claude ./cmd/claude

# Clean build artifacts
clean:
	rm -rf bin/

# Run the CLI
run: build
	./bin/claude

# Run tests
test:
	go test ./...

# Install to GOPATH/bin
install:
	go install ./cmd/claude

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy
