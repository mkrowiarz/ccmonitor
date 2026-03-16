APP := ccmonitor
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
PREFIX ?= /usr/local

.PHONY: build run test clean install uninstall release

build:
	go build -ldflags "$(LDFLAGS)" -o $(APP) .

run: build
	./$(APP)

test:
	go test ./...

clean:
	rm -f $(APP)
	rm -rf dist/

install: build
	install -d $(PREFIX)/bin
	install -m 755 $(APP) $(PREFIX)/bin/$(APP)

uninstall:
	rm -f $(PREFIX)/bin/$(APP)

release:
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP)-darwin-amd64 .
	GOOS=linux  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP)-linux-amd64  .
	GOOS=linux  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP)-linux-arm64  .
	@echo "Built binaries in dist/"
	@ls -lh dist/
