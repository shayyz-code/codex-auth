BINARY := bin/codex-auth
PACKAGE_TESTS := npm/*.test.js
VERSION ?= dev
OUT ?= $(BINARY)

.PHONY: build check clean release-build test test-go test-npm version

build:
	go build -o $(BINARY) ./cmd/codex-auth

check: test build

clean:
	rm -rf bin dist

release-build:
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o "$(OUT)" ./cmd/codex-auth

version:
	npm run version:update -- "$(VERSION)"

test: test-go test-npm

test-go:
	go test ./...

test-npm:
	node --test $(PACKAGE_TESTS)
