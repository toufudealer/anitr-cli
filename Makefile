GO=go
BINARY_NAME=anitr-cli
BUILD_DIR=./build

INSTALL_DIR_WINDOWS="C:/Program Files/anitr-cli"

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILDENV := $(shell go version)

LDFLAGS=-ldflags="-X 'github.com/xeyossr/anitr-cli/internal/update.version=$(VERSION)' -X 'github.com/xeyossr/anitr-cli/internal/update.buildEnv=$(BUILDENV)'"

mod-tidy:
	$(GO) mod tidy

build-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-x86_64.exe

build: mod-tidy
	mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)

build-all: mod-tidy build-windows

install-linux: build
	chmod +x $(BUILD_DIR)/$(BINARY_NAME)
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR_LINUX)/$(BINARY_NAME)

install-windows: build
	powershell -Command "New-Item -ItemType Directory -Force -Path $(INSTALL_DIR_WINDOWS)"
	powershell -Command "Copy-Item -Path $(BUILD_DIR)/$(BINARY_NAME) -Destination $(INSTALL_DIR_WINDOWS)/$(BINARY_NAME).exe -Force"

install-macos: build
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR_MAC)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)

all: build-windows install-windows
