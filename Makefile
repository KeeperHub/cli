VERSION ?= $(shell python3 -c "import json; print(json.load(open('.release-please-manifest.json'))['.'])" 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/keeperhub/cli/internal/version.Version=$(VERSION)

.PHONY: build test lint clean install sync-version

sync-version:
	@cp .release-please-manifest.json internal/version/manifest.json

build: sync-version
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/kh ./cmd/kh

test:
	go test -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

install: sync-version
	CGO_ENABLED=0 go install -ldflags="$(LDFLAGS)" ./cmd/kh
