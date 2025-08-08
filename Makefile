GO=go
BINARY_NAME=anitr-cli
BUILD_DIR=./build
INSTALL_DIR=/usr/bin

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILDENV := $(shell go version)

LDFLAGS=-ldflags="-X 'github.com/xeyossr/anitr-cli/internal/update.version=$(VERSION)' -X 'github.com/xeyossr/anitr-cli/internal/update.buildEnv=$(BUILDENV)'"

mod-tidy:
	$(GO) mod tidy

build: mod-tidy
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

install: build
	chmod +x $(BUILD_DIR)/$(BINARY_NAME)
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)

all: mod-tidy build install