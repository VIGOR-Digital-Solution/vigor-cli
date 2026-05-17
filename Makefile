.PHONY: build test lint fmt run snapshot release install clean

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	go build -trimpath -ldflags="$(LDFLAGS)" -o bin/vigor ./cmd/vigor

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

fmt:
	gofmt -w -s .
	goimports -w .

run: build
	./bin/vigor $(ARGS)

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean

install: build
	cp bin/vigor $$(go env GOPATH)/bin/vigor

clean:
	rm -rf bin/ dist/
