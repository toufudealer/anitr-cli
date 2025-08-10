GO=go
BINARY_NAME=anitr-cli
BUILD_DIR=./build
BUILD_DIR_WIN=$(BUILD_DIR)/windows
BUILD_DIR_LINUX=$(BUILD_DIR)/linux
BUILD_DIR_MAC=$(BUILD_DIR)/macos

INSTALL_DIR_LINUX=/usr/bin
INSTALL_DIR_WINDOWS="C:/Program Files/anitr-cli"
INSTALL_DIR_MAC=/usr/local/bin

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILDENV := $(shell go version)

LDFLAGS=-ldflags="-X 'github.com/xeyossr/anitr-cli/internal/update.version=$(VERSION)' -X 'github.com/xeyossr/anitr-cli/internal/update.buildEnv=$(BUILDENV)'"

mod-tidy:
	$(GO) mod tidy

build-linux:
	mkdir -p $(BUILD_DIR_LINUX)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR_LINUX)/$(BINARY_NAME)

build-windows:
	mkdir -p $(BUILD_DIR_WIN)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR_WIN)/$(BINARY_NAME).exe

build-macos:
	mkdir -p $(BUILD_DIR_MAC)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR_MAC)/$(BINARY_NAME)-amd64
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR_MAC)/$(BINARY_NAME)-arm64

build: mod-tidy build-linux build-windows build-macos

install-linux: build-linux
	chmod +x $(BUILD_DIR_LINUX)/$(BINARY_NAME)
	sudo mv $(BUILD_DIR_LINUX)/$(BINARY_NAME) $(INSTALL_DIR_LINUX)/$(BINARY_NAME)

install-windows: build-windows
	powershell -Command "New-Item -ItemType Directory -Force -Path $(INSTALL_DIR_WINDOWS)"
	powershell -Command "Copy-Item -Path $(BUILD_DIR_WIN)/$(BINARY_NAME).exe -Destination $(INSTALL_DIR_WINDOWS)/$(BINARY_NAME).exe -Force"

install-macos: build-macos
	@echo "Mac işlemci mimarisi seçin:"
	@echo "1) amd64 (Intel)"
	@echo "2) arm64 (Apple Silicon)"
	@read -p "Seçiminiz (1/2): " choice; \
	if [ "$$choice" = "1" ]; then \
		sudo mv $(BUILD_DIR_MAC)/$(BINARY_NAME)-amd64 $(INSTALL_DIR_MAC)/$(BINARY_NAME); \
	elif [ "$$choice" = "2" ]; then \
		sudo mv $(BUILD_DIR_MAC)/$(BINARY_NAME)-arm64 $(INSTALL_DIR_MAC)/$(BINARY_NAME); \
	else \
		echo "Geçersiz seçim! Kurulum iptal edildi."; \
		exit 1; \
	fi

clean:
	rm -rf $(BUILD_DIR)

all: build install
