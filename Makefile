APP     := wps365-oauth-cli
BUILD   := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all linux windows macos clean

all: linux windows macos

linux:
	@mkdir -p $(BUILD)/linux
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD)/linux/$(APP)-amd64   .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD)/linux/$(APP)-arm64   .

windows:
	@mkdir -p $(BUILD)/windows
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD)/windows/$(APP)-amd64.exe .

macos:
	@mkdir -p $(BUILD)/macos
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD)/macos/$(APP)-amd64  .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD)/macos/$(APP)-arm64  .

clean:
	rm -rf $(BUILD)
