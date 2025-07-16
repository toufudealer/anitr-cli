# Makefile

GO=go
BINARY_NAME=anitr-cli
INSTALL_DIR=/usr/bin

build:
	$(GO) build -o $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

install: build
	sudo mv $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	echo "Uygulama başarıyla /usr/bin/ dizinine yüklendi."

clean:
	rm -f $(BINARY_NAME)

all: build install
