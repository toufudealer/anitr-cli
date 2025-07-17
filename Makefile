GO=go
BINARY_NAME=anitr-cli
INSTALL_DIR=/usr/bin

mod-tidy:
	$(GO) mod tidy

build: mod-tidy
	$(GO) build -o $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

install: build
	chmod +x $(BINARY_NAME)
	sudo mv $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

all: mod-tidy build install
