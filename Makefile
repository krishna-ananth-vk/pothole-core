# Makefile for Go backend builds

APP_NAME := backend
BUILD_DIR := bin

# Default target
.PHONY: all build-linux build-macos run clean

all: build-linux

build-linux:
	@echo "Building $(APP_NAME) for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux .

build-macos:
	@echo "Building $(APP_NAME) for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-macos .

run:
	@echo "Running $(APP_NAME)..."
	go run .

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
