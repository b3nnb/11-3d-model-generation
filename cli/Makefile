.PHONY: build install clean

BINARY = modelgen
INSTALL_DIR = $(HOME)/.local/bin

build:
	go build -o $(BINARY) .

install: build
	install -d $(INSTALL_DIR)
	install -m 0755 $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "✅ Installed: $(INSTALL_DIR)/$(BINARY)"
	@echo "   Add to ~/.bashrc:  export PATH=\"\$$HOME/.local/bin:\$$PATH\""

clean:
	rm -f $(BINARY)

# Cross-compile for macOS ARM (mini)
build-darwin:
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY)-darwin-arm64 .

# Cross-compile for Linux AMD64 (friday server)
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 .

# Build both platforms
build-all: build-darwin build-linux
	@echo "Built: $(BINARY)-darwin-arm64  $(BINARY)-linux-amd64"
