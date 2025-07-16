GO=go
BINARY_NAME=anitr-cli
INSTALL_DIR=/usr/bin

mod-init:
	$(GO) mod init

mod-tidy:
	$(GO) mod tidy

build: mod-tidy
	$(GO) build -o $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

install: build
	sudo mv $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	echo "Uygulama başarıyla /usr/bin/ dizinine yüklendi."

clean:
	rm -f $(BINARY_NAME)

all: mod-tidy build install
