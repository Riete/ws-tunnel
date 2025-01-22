# Makefile

BINARY_NAME := ws-tunnel
BUILD_IMAGE := ghcr.io/riete/golang:1.23.4-busybox

.PHONY: build-linux-amd64
build-linux-amd64:
	docker run --rm -w /ws-tunnel --platform linux/amd64 -v .:/ws-tunnel $(BUILD_IMAGE) go build -ldflags="-s -w" -o $(BINARY_NAME)-amd64
	upx $(BINARY_NAME)-amd64

.PHONY: build-linux-arm64
build-linux-arm64:
	docker run --rm -w /ws-tunnel --platform linux/arm64 -v .:/ws-tunnel $(BUILD_IMAGE) go build -ldflags="-s -w" -o $(BINARY_NAME)-arm64
	upx $(BINARY_NAME)-arm64

.PHONY: build-linux
build-linux: build-linux-amd64 build-linux-arm64

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)-amd64 $(BINARY_NAME)-arm64
