VERSION ?= dev
LDFLAGS := -s -w -X github.com/keeperhub/cli/internal/version.Version=$(VERSION)

.PHONY: build test lint clean install

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/kh ./cmd/kh

test:
	go test -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

install:
	CGO_ENABLED=0 go install -ldflags="$(LDFLAGS)" ./cmd/kh
