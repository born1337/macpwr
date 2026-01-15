# macpwr Makefile (Go version)

BINARY_NAME=macpwr
PREFIX ?= /usr/local
BIN_DIR = $(PREFIX)/bin

.PHONY: all build install uninstall clean test bench

all: build

build:
	@echo "Building macpwr..."
	@go build -o $(BINARY_NAME) ./cmd/macpwr
	@echo "✓ Built $(BINARY_NAME)"

install: build
	@echo "Installing macpwr to $(BIN_DIR)..."
	@mkdir -p $(BIN_DIR)
	@cp $(BINARY_NAME) $(BIN_DIR)/$(BINARY_NAME)
	@chmod +x $(BIN_DIR)/$(BINARY_NAME)
	@echo "✓ Installed successfully!"
	@echo ""
	@echo "Run 'macpwr help' to get started."

uninstall:
	@echo "Uninstalling macpwr..."
	@rm -f $(BIN_DIR)/$(BINARY_NAME)
	@echo "✓ Uninstalled successfully!"

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean
	@echo "✓ Cleaned"

test:
	@echo "Running tests..."
	@go test ./...

bench:
	@echo "Running benchmarks..."
	@echo ""
	@echo "=== Status command ==="
	@time ./$(BINARY_NAME) > /dev/null 2>&1
	@time ./$(BINARY_NAME) > /dev/null 2>&1
	@time ./$(BINARY_NAME) > /dev/null 2>&1
	@echo ""
	@echo "=== Show command ==="
	@time ./$(BINARY_NAME) show > /dev/null 2>&1
	@time ./$(BINARY_NAME) show > /dev/null 2>&1
	@time ./$(BINARY_NAME) show > /dev/null 2>&1
	@echo ""
	@echo "=== Battery command ==="
	@time ./$(BINARY_NAME) battery > /dev/null 2>&1
	@time ./$(BINARY_NAME) battery > /dev/null 2>&1
	@time ./$(BINARY_NAME) battery > /dev/null 2>&1
	@echo ""
	@echo "=== Help command ==="
	@time ./$(BINARY_NAME) help > /dev/null 2>&1
	@time ./$(BINARY_NAME) help > /dev/null 2>&1
	@time ./$(BINARY_NAME) help > /dev/null 2>&1

help:
	@echo "macpwr Makefile (Go version)"
	@echo ""
	@echo "Usage:"
	@echo "  make build      Build the binary"
	@echo "  make install    Install macpwr to $(PREFIX)/bin"
	@echo "  make uninstall  Remove macpwr from $(PREFIX)/bin"
	@echo "  make clean      Remove build artifacts"
	@echo "  make test       Run tests"
	@echo "  make bench      Run benchmarks"
	@echo "  make help       Show this help"
	@echo ""
	@echo "Options:"
	@echo "  PREFIX=/path    Installation prefix (default: /usr/local)"
