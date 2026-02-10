VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
MODULE := github.com/Mgrdich/TermChess
LDFLAGS := -s -w \
    -X $(MODULE)/internal/version.Version=$(VERSION) \
    -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE) \
    -X $(MODULE)/internal/version.GitCommit=$(GIT_COMMIT)

.PHONY: build build-all checksums test run clean

build:
	go build -ldflags="$(LDFLAGS)" -o bin/termchess ./cmd/termchess

build-all:
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/termchess-$(VERSION)-darwin-amd64 ./cmd/termchess
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/termchess-$(VERSION)-darwin-arm64 ./cmd/termchess
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/termchess-$(VERSION)-linux-amd64 ./cmd/termchess
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/termchess-$(VERSION)-linux-arm64 ./cmd/termchess

checksums:
	@cd dist && (command -v sha256sum >/dev/null 2>&1 && sha256sum termchess-* > checksums.txt || shasum -a 256 termchess-* > checksums.txt)

test:
	go test -v ./...

run:
	go run ./cmd/termchess

clean:
	rm -rf bin/ dist/
