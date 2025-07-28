GO=go
BINARY_NAME=anitr-cli
BUILD_DIR=./build
INSTALL_DIR=/usr/bin

mod-tidy:
	$(GO) mod tidy

build: mod-tidy
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)

run: build
	./$(BINARY_NAME)

install: build
	chmod +x $(BUILD_DIR)/$(BINARY_NAME)
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)

all: mod-tidy build install
