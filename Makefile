GO=go
BINARY_NAME=anitr-cli
BUILD_DIR=./build

INSTALL_DIR_LINUX=/usr/bin
INSTALL_DIR_WINDOWS="C:/Program Files/anitr-cli"
INSTALL_DIR_MAC=/usr/local/bin

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILDENV := $(shell go version)

LDFLAGS=-ldflags="-X 'github.com/xeyossr/anitr-cli/internal/update.version=$(VERSION)' -X 'github.com/xeyossr/anitr-cli/internal/update.buildEnv=$(BUILDENV)'"

mod-tidy:
	$(GO) mod tidy

build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-x86_64
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64

build-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-x86_64.exe

build-macos:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64

build: mod-tidy build-linux build-windows build-macos

install-linux: build-linux
	@echo "Linux işlemci mimarisi seçin:"
	@echo "1) amd64 (x86_64)"
	@echo "2) arm64 (aarch64)"
	@read -p "Seçiminiz (1/2): " choice; \
	if [ "$$choice" = "1" ]; then \
		chmod +x $(BUILD_DIR)/$(BINARY_NAME)-linux-x86_64; \
		sudo mv $(BUILD_DIR)/$(BINARY_NAME)-linux-x86_64 $(INSTALL_DIR_LINUX)/$(BINARY_NAME); \
	elif [ "$$choice" = "2" ]; then \
		chmod +x $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64; \
		sudo mv $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(INSTALL_DIR_LINUX)/$(BINARY_NAME); \
	else \
		echo "Geçersiz seçim! Kurulum iptal edildi."; \
		exit 1; \
	fi

install-windows: build-windows
	powershell -Command "New-Item -ItemType Directory -Force -Path $(INSTALL_DIR_WINDOWS)"
	powershell -Command "Copy-Item -Path $(BUILD_DIR)/$(BINARY_NAME)-windows-x86_64.exe -Destination $(INSTALL_DIR_WINDOWS)/$(BINARY_NAME).exe -Force"

install-macos: build-macos
	@echo "Mac işlemci mimarisi seçin:"
	@echo "1) amd64 (Intel)"
	@echo "2) arm64 (Apple Silicon)"
	@read -p "Seçiminiz (1/2): " choice; \
	if [ "$$choice" = "1" ]; then \
		sudo mv $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 $(INSTALL_DIR_MAC)/$(BINARY_NAME); \
	elif [ "$$choice" = "2" ]; then \
		sudo mv $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 $(INSTALL_DIR_MAC)/$(BINARY_NAME); \
	else \
		echo "Geçersiz seçim! Kurulum iptal edildi."; \
		exit 1; \
	fi

clean:
	rm -rf $(BUILD_DIR)

all: build install
