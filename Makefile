APP_NAME := opencode-stats
BUILD_DIR := bin
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: build install clean test

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) .

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(APP_NAME) $(INSTALL_DIR)/$(APP_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(APP_NAME)"

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./... -v -cover

run: build
	./$(BUILD_DIR)/$(APP_NAME)
